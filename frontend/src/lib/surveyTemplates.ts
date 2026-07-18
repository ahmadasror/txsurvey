import type { QuestionInput, QuestionType } from "@/types/forms";

export interface SurveyTemplate {
  id: string;
  publicPath: string;
  theme: string;
  font: string;
  title: string;
  seoTitle: string;
  description: string;
  welcomeDesc: string;
  audience: string;
  questions: QuestionInput[];
}

const choice = (title: string, labels: string[], required = true): QuestionInput => ({
  type: "multiple_choice",
  title,
  required,
  metadata: { options: labels.map((label) => ({ id: "", label })) },
});
const rating = (title: string, scale: number): QuestionInput => ({
  type: "rating",
  title,
  required: true,
  metadata: { scale },
});
const yesno = (title: string): QuestionInput => ({ type: "yes_no", title, required: true, metadata: {} });
const long = (title: string, required = false): QuestionInput => ({ type: "long_text", title, required, metadata: {} });

export const SURVEY_TEMPLATES: SurveyTemplate[] = [
  {
    id: "pulse",
    publicPath: "template-survei-kepuasan-karyawan",
    theme: "pine",
    font: "editorial",
    title: "Pulse Karyawan",
    seoTitle: "Template Survei Kepuasan Karyawan",
    description: "Ukur kepuasan, loyalitas, dan area perbaikan tim dengan lima pertanyaan singkat.",
    welcomeDesc: "Beberapa pertanyaan singkat soal pengalaman kerjamu. Anonim.",
    audience: "HR, People Operations, dan pemimpin tim yang ingin memantau kesehatan organisasi secara rutin.",
    questions: [
      choice("Sudah berapa lama kamu bergabung?", ["< 6 bulan", "6–12 bulan", "1–3 tahun", "> 3 tahun"]),
      choice("Seberapa puas kamu bekerja di sini?", ["Sangat puas", "Puas", "Biasa aja", "Kurang puas", "Tidak puas"]),
      rating("Seberapa besar kemungkinan kamu merekomendasikan tempat ini ke teman?", 10),
      long("Apa satu hal yang bisa kami perbaiki?"),
      yesno("Boleh kami follow-up jawabanmu?"),
    ],
  },
  {
    id: "onboarding",
    publicPath: "template-feedback-onboarding",
    theme: "sand",
    font: "editorial",
    title: "Feedback Onboarding",
    seoTitle: "Template Survei Feedback Onboarding",
    description: "Pahami pengalaman karyawan baru dan temukan bagian onboarding yang perlu diperbaiki.",
    welcomeDesc: "Bantu kami memperbaiki proses onboarding untuk yang berikutnya.",
    audience: "Tim HR dan hiring manager yang mengevaluasi pengalaman 30–90 hari pertama karyawan.",
    questions: [
      rating("Seberapa jelas peran & tanggung jawabmu setelah onboarding?", 5),
      choice("Apakah pelatihan yang diberikan cukup?", ["Lebih dari cukup", "Cukup", "Kurang", "Tidak ada pelatihan"]),
      yesno("Apakah mentormu membantu?"),
      long("Bagian onboarding mana yang paling membantu?"),
      long("Apa yang sebaiknya kami tambahkan?"),
    ],
  },
  {
    id: "it-csat",
    publicPath: "template-kepuasan-layanan-it",
    theme: "ink",
    font: "soft",
    title: "Kepuasan Layanan IT",
    seoTitle: "Template Survei Kepuasan Layanan IT",
    description: "Nilai penyelesaian masalah, kecepatan, dan kualitas dukungan tim IT internal.",
    welcomeDesc: "Ceritakan pengalamanmu dengan layanan IT terakhir kali.",
    audience: "Helpdesk dan tim IT internal yang ingin mengukur CSAT setelah tiket selesai.",
    questions: [
      choice("Layanan IT mana yang kamu gunakan?", ["Helpdesk", "Perbaikan perangkat", "Akses & akun", "Jaringan", "Lainnya"]),
      yesno("Apakah masalahmu terselesaikan?"),
      rating("Seberapa cepat penanganannya?", 5),
      choice("Secara keseluruhan, seberapa puas kamu?", ["Sangat puas", "Puas", "Biasa aja", "Tidak puas"]),
      long("Saran untuk tim IT?"),
    ],
  },
];

export const templateByPath = (path?: string) => SURVEY_TEMPLATES.find((template) => template.publicPath === path);

export const uniqueTemplateTypes = (questions: QuestionInput[]): QuestionType[] => [
  ...new Set(questions.map((question) => question.type)),
];
