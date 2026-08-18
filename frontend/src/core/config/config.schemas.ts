import { z } from "zod";

const SentrySchema = z.object({
  dsn: z.string(),
  environment: z.string(),
  version: z.string(),
  traces_sample_rate: z.number().min(0).max(1).default(1),
  replay_session_sample_rate: z.number().min(0).max(1).default(0),
  replay_error_sample_rate: z.number().min(0).max(1).default(1),
});

const DateFormatSchema = z.object({
  locale: z.string(),
  options: z.record(z.string(), z.any()).default({}),
});

export const PlaylistStepSchema = z.object({
  type: z.string(),
  order: z.enum(["random", "asc", "desc"]),
  count: z.preprocess(
    (v) => (v === 0 ? undefined : v),
    z.number().int().positive().default(1),
  ),
  duration: z.preprocess(
    (v) => (v === 0 ? undefined : v),
    z.number().int().positive().default(10),
  ),
});

export const PlaylistSchema = z.object({
  id: z.number(),
  name: z.string(),
  steps: z.array(PlaylistStepSchema),
});

export const ExternalAdminURLsSchema = z.object({
  timetable: z.url().or(z.literal("")).optional(),
  news: z.url().or(z.literal("")).optional(),
});

export const BeamerConfigSchema = z.object({
  allowed_animations: z
    .array(z.enum(["fade", "slideRight", "zoomIn", "flip", "urgent"]))
    .default(["fade", "slideRight", "zoomIn", "flip", "urgent"]),
});

export const AppConfigSchema = z.object({
  version: z.string(),
  environment: z.string(),
  environment_message: z.string().optional(),
  sentry: SentrySchema,
  format: z.object({
    date: DateFormatSchema,
  }),
  playlists: z
    .array(PlaylistSchema)
    .min(1, "At least one playlist must be defined"),
  admin_urls: ExternalAdminURLsSchema,
  beamer: BeamerConfigSchema.default({
    allowed_animations: ["fade", "slideRight", "zoomIn", "flip", "urgent"],
  }),
  auth: z
    .object({
      type: z.enum(["oidc"]),
      name: z.string().min(1),
      authority: z.url(),
      client_id: z.string().min(1),
    })
    .optional(),
});

export type AppConfig = z.infer<typeof AppConfigSchema>;
export type SentryConfig = z.infer<typeof SentrySchema>;
export type Playlist = z.infer<typeof PlaylistSchema>;
export type PlaylistStep = z.infer<typeof PlaylistStepSchema>;
export type BeamerConfig = z.infer<typeof BeamerConfigSchema>;
