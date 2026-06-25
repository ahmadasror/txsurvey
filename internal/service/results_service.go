package service

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/ahmadasror/txsurvey/internal/model"
	"github.com/ahmadasror/txsurvey/internal/repository"
)

// ResultsService serves a form owner's results: response list/detail, analytics,
// and CSV export. Every method verifies form ownership first.
type ResultsService struct {
	forms     *repository.FormRepo
	questions *repository.QuestionRepo
	responses *repository.ResponseRepo
}

func NewResultsService(forms *repository.FormRepo, questions *repository.QuestionRepo, responses *repository.ResponseRepo) *ResultsService {
	return &ResultsService{forms: forms, questions: questions, responses: responses}
}

func (s *ResultsService) requireForm(ctx context.Context, ownerID, formID string) (*model.Form, error) {
	form, err := s.forms.GetByIDOwned(ctx, formID, ownerID)
	if err != nil {
		return nil, err
	}
	if form == nil {
		return nil, errFormNotFound
	}
	return form, nil
}

// ListResponses returns a page of responses with answers, plus the total.
func (s *ResultsService) ListResponses(ctx context.Context, ownerID, formID string, page, perPage int) ([]model.Response, int, error) {
	if _, err := s.requireForm(ctx, ownerID, formID); err != nil {
		return nil, 0, err
	}
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 200 {
		perPage = 50
	}
	return s.responses.ListByForm(ctx, formID, perPage, (page-1)*perPage)
}

// GetResponse returns one response with its answers.
func (s *ResultsService) GetResponse(ctx context.Context, ownerID, formID, responseID string) (*model.Response, error) {
	if _, err := s.requireForm(ctx, ownerID, formID); err != nil {
		return nil, err
	}
	resp, err := s.responses.GetByID(ctx, formID, responseID)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, errFormNotFound
	}
	return resp, nil
}

// Analytics aggregates per-question summaries plus response/completion totals.
func (s *ResultsService) Analytics(ctx context.Context, ownerID, formID string) (*model.FormAnalytics, error) {
	if _, err := s.requireForm(ctx, ownerID, formID); err != nil {
		return nil, err
	}
	questions, err := s.questions.ListByForm(ctx, formID)
	if err != nil {
		return nil, err
	}
	answers, err := s.responses.AllAnswers(ctx, formID)
	if err != nil {
		return nil, err
	}
	total, err := s.responses.CountByForm(ctx, formID)
	if err != nil {
		return nil, err
	}
	completed, err := s.responses.CompletedCountByForm(ctx, formID)
	if err != nil {
		return nil, err
	}

	byQuestion := make(map[string][]json.RawMessage)
	for _, a := range answers {
		byQuestion[a.QuestionID] = append(byQuestion[a.QuestionID], a.Value)
	}

	summaries := make([]model.QuestionSummary, 0, len(questions))
	for _, q := range questions {
		if q.Type == model.QStatement {
			continue
		}
		summaries = append(summaries, computeSummary(q, byQuestion[q.ID]))
	}

	rate := 0.0
	if total > 0 {
		rate = float64(completed) / float64(total)
	}
	return &model.FormAnalytics{ResponseCount: total, CompletionRate: rate, Questions: summaries}, nil
}

// computeSummary builds a single question's analytics from its raw answers.
func computeSummary(q model.Question, raws []json.RawMessage) model.QuestionSummary {
	sum := model.QuestionSummary{QuestionID: q.ID, Title: q.Title, Type: q.Type, Answered: len(raws)}

	switch q.Type {
	case model.QMultipleChoice, model.QDropdown, model.QCheckboxes:
		counts := map[string]int{}
		for _, raw := range raws {
			if q.Type == model.QCheckboxes {
				var arr []string
				if json.Unmarshal(raw, &arr) == nil {
					for _, id := range arr {
						counts[id]++
					}
				}
			} else {
				if s, err := decodeString(raw); err == nil {
					counts[s]++
				}
			}
		}
		for _, o := range q.Metadata.Options {
			sum.Options = append(sum.Options, model.OptionCount{Value: o.ID, Label: o.Label, Count: counts[o.ID]})
		}

	case model.QYesNo:
		var yes, no int
		for _, raw := range raws {
			if b, err := decodeBool(raw); err == nil {
				if b {
					yes++
				} else {
					no++
				}
			}
		}
		sum.Options = []model.OptionCount{
			{Value: "true", Label: "Yes", Count: yes},
			{Value: "false", Label: "No", Count: no},
		}

	case model.QRating:
		scale := q.Metadata.Scale
		if scale == 0 {
			scale = 5
		}
		counts := make([]int, scale+1)
		var total float64
		for _, raw := range raws {
			if f, err := decodeNumber(raw); err == nil {
				n := int(f)
				if n >= 1 && n <= scale {
					counts[n]++
					total += f
				}
			}
		}
		for n := 1; n <= scale; n++ {
			sum.Options = append(sum.Options, model.OptionCount{Value: strconv.Itoa(n), Label: strconv.Itoa(n), Count: counts[n]})
		}
		if len(raws) > 0 {
			avg := total / float64(len(raws))
			sum.Average = &avg
		}

	case model.QNumber:
		if len(raws) > 0 {
			var total float64
			var n int
			for _, raw := range raws {
				if f, err := decodeNumber(raw); err == nil {
					total += f
					n++
				}
			}
			if n > 0 {
				avg := total / float64(n)
				sum.Average = &avg
			}
		}
	}
	return sum
}

// ExportCSV writes all responses as CSV (one column per answerable question) and
// returns a suggested filename.
func (s *ResultsService) ExportCSV(ctx context.Context, ownerID, formID string, out io.Writer) (string, error) {
	form, err := s.requireForm(ctx, ownerID, formID)
	if err != nil {
		return "", err
	}
	questions, err := s.questions.ListByForm(ctx, formID)
	if err != nil {
		return "", err
	}
	// All responses (small scale — a generous single page).
	responses, _, err := s.responses.ListByForm(ctx, formID, 100000, 0)
	if err != nil {
		return "", err
	}

	answerable := make([]model.Question, 0, len(questions))
	for _, q := range questions {
		if q.Type != model.QStatement {
			answerable = append(answerable, q)
		}
	}

	w := csv.NewWriter(out)
	header := []string{"Response ID", "Submitted At"}
	for _, q := range answerable {
		title := q.Title
		if title == "" {
			title = "(untitled)"
		}
		header = append(header, title)
	}
	if err := w.Write(header); err != nil {
		return "", err
	}

	for _, resp := range responses {
		byQuestion := make(map[string]json.RawMessage, len(resp.Answers))
		for _, a := range resp.Answers {
			byQuestion[a.QuestionID] = a.Value
		}
		row := []string{resp.ID, resp.SubmittedAt.UTC().Format(time.RFC3339)}
		for _, q := range answerable {
			row = append(row, formatAnswer(q, byQuestion[q.ID]))
		}
		if err := w.Write(row); err != nil {
			return "", err
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return "", err
	}
	return form.Slug + "-responses.csv", nil
}

// formatAnswer renders an answer value as a human-readable CSV cell.
func formatAnswer(q model.Question, raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	labelByID := map[string]string{}
	for _, o := range q.Metadata.Options {
		labelByID[o.ID] = o.Label
	}

	switch q.Type {
	case model.QMultipleChoice, model.QDropdown:
		if s, err := decodeString(raw); err == nil {
			if lbl, ok := labelByID[s]; ok {
				return lbl
			}
			return s
		}
	case model.QCheckboxes:
		var arr []string
		if json.Unmarshal(raw, &arr) == nil {
			labels := make([]string, 0, len(arr))
			for _, id := range arr {
				if lbl, ok := labelByID[id]; ok {
					labels = append(labels, lbl)
				} else {
					labels = append(labels, id)
				}
			}
			return strings.Join(labels, "; ")
		}
	case model.QYesNo:
		if b, err := decodeBool(raw); err == nil {
			if b {
				return "Yes"
			}
			return "No"
		}
	case model.QRating, model.QNumber:
		if f, err := decodeNumber(raw); err == nil {
			return strconv.FormatFloat(f, 'f', -1, 64)
		}
	}
	// text / email / date and fallbacks: decode as string, else raw JSON.
	if s, err := decodeString(raw); err == nil {
		return s
	}
	return string(raw)
}
