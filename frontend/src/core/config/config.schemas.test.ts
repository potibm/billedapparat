import { describe, expect, it } from "vitest";
import {
  AppConfigSchema,
  PlaylistSchema,
  PlaylistStepSchema,
  ExternalAdminURLsSchema,
} from "./config.schemas";

describe("config.schemas", () => {
  describe("PlaylistStepSchema", () => {
    it("should parse a valid step with defaults", () => {
      const result = PlaylistStepSchema.safeParse({
        type: "sponsor",
        order: "random",
      });
      expect(result.success).toBe(true);
      if (result.success) {
        expect(result.data.type).toBe("sponsor");
        expect(result.data.order).toBe("random");
        expect(result.data.count).toBe(1);
        expect(result.data.duration).toBe(10);
      }
    });

    it("should treat count=0 as undefined and fall back to default", () => {
      const result = PlaylistStepSchema.safeParse({
        type: "news",
        order: "asc",
        count: 0,
        duration: 0,
      });
      expect(result.success).toBe(true);
      if (result.success) {
        expect(result.data.count).toBe(1);
        expect(result.data.duration).toBe(10);
      }
    });

    it("should reject an invalid order value", () => {
      const result = PlaylistStepSchema.safeParse({
        type: "news",
        order: "invalid",
      });
      expect(result.success).toBe(false);
    });

    it("should reject a negative count", () => {
      const result = PlaylistStepSchema.safeParse({
        type: "news",
        order: "asc",
        count: -1,
      });
      expect(result.success).toBe(false);
    });
  });

  describe("PlaylistSchema", () => {
    it("should parse a valid playlist", () => {
      const result = PlaylistSchema.safeParse({
        id: 1,
        name: "Default",
        steps: [{ type: "sponsor", order: "random" }],
      });
      expect(result.success).toBe(true);
    });

    it("should allow empty name (schema does not enforce min length)", () => {
      const result = PlaylistSchema.safeParse({
        id: 1,
        name: "",
        steps: [{ type: "sponsor", order: "random" }],
      });
      expect(result.success).toBe(true);
    });
  });

  describe("ExternalAdminURLsSchema", () => {
    it("should parse valid URLs", () => {
      const result = ExternalAdminURLsSchema.safeParse({
        timetable: "https://timetable.example.com",
        news: "https://news.example.com",
      });
      expect(result.success).toBe(true);
    });

    it("should allow empty strings", () => {
      const result = ExternalAdminURLsSchema.safeParse({
        timetable: "",
        news: "",
      });
      expect(result.success).toBe(true);
    });

    it("should reject invalid URLs", () => {
      const result = ExternalAdminURLsSchema.safeParse({
        timetable: "not-a-url",
      });
      expect(result.success).toBe(false);
    });

    it("should parse when keys are missing", () => {
      const result = ExternalAdminURLsSchema.safeParse({});
      expect(result.success).toBe(true);
    });
  });

  describe("AppConfigSchema", () => {
    const validBase = {
      version: "1.0.0",
      environment: "production",
      sentry: {
        dsn: "https://key@sentry.io/1",
        environment: "production",
        version: "1.0.0",
      },
      format: {
        date: {
          locale: "da-DK",
          options: { weekday: "long" },
        },
      },
      playlists: [
        {
          id: 1,
          name: "Default",
          steps: [{ type: "sponsor", order: "random" }],
        },
      ],
      admin_urls: {
        timetable: "",
        news: "",
      },
    };

    it("should parse a minimal valid config", () => {
      const result = AppConfigSchema.safeParse(validBase);
      expect(result.success).toBe(true);
    });

    it("should reject a config with no playlists", () => {
      const result = AppConfigSchema.safeParse({
        ...validBase,
        playlists: [],
      });
      expect(result.success).toBe(false);
    });

    it("should parse a config with optional auth", () => {
      const result = AppConfigSchema.safeParse({
        ...validBase,
        auth: {
          type: "oidc",
          name: "Dex",
          authority: "https://dex.example.com",
          client_id: "billedapparat",
        },
      });
      expect(result.success).toBe(true);
    });

    it("should reject a config with invalid auth authority", () => {
      const result = AppConfigSchema.safeParse({
        ...validBase,
        auth: {
          type: "oidc",
          name: "Dex",
          authority: "not-a-url",
          client_id: "billedapparat",
        },
      });
      expect(result.success).toBe(false);
    });

    it("should include default sentry sample rates", () => {
      const result = AppConfigSchema.safeParse(validBase);
      expect(result.success).toBe(true);
      if (result.success) {
        expect(result.data.sentry.traces_sample_rate).toBe(1);
        expect(result.data.sentry.replay_session_sample_rate).toBe(0);
        expect(result.data.sentry.replay_error_sample_rate).toBe(1);
      }
    });

    it("should reject invalid sentry sample rates", () => {
      const result = AppConfigSchema.safeParse({
        ...validBase,
        sentry: {
          ...validBase.sentry,
          traces_sample_rate: 1.5,
        },
      });
      expect(result.success).toBe(false);
    });
  });
});
