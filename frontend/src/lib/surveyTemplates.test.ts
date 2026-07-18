import { describe, expect, it } from "vitest";
import { SURVEY_TEMPLATES, templateByPath } from "./surveyTemplates";

describe("public survey templates", () => {
  it("keeps ids and public paths unique", () => {
    expect(new Set(SURVEY_TEMPLATES.map((template) => template.id)).size).toBe(SURVEY_TEMPLATES.length);
    expect(new Set(SURVEY_TEMPLATES.map((template) => template.publicPath)).size).toBe(SURVEY_TEMPLATES.length);
  });

  it("publishes complete previews backed by usable product templates", () => {
    for (const template of SURVEY_TEMPLATES) {
      expect(template.title).not.toBe("");
      expect(template.description.length).toBeGreaterThan(40);
      expect(template.questions.length).toBeGreaterThanOrEqual(5);
      expect(templateByPath(template.publicPath)).toBe(template);
    }
  });
});
