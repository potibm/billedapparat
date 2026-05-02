import { z } from "zod";

const SentrySchema = z.object({
  dsn: z.string(),
  environment: z.string(),
  version: z.string(),
});

const DateFormatSchema = z.object({
  locale: z.string(),
  options: z.record(z.string(), z.any()).default({}),
});

const PlaylistStepSchema = z.object({
  type: z.string(),
  order: z.enum(["random", "asc", "desc"]),
  count: z.number().positive().default(1),
  duration: z.number().positive().default(10),
});

const PlaylistSchema = z.object({
  id: z.number(),
  name: z.string(),
  steps: z.array(PlaylistStepSchema),
});

export const AppConfigSchema = z.object({
  version: z.string(),
  environment: z.string(),
  environment_message: z.string().optional(),
  sentry: SentrySchema,
  format: z.object({
    date: DateFormatSchema,
  }),
  playlists: z.array(PlaylistSchema),
});

export type AppConfig = z.infer<typeof AppConfigSchema>;
export type Playlist = z.infer<typeof PlaylistSchema>;
export type PlaylistStep = z.infer<typeof PlaylistStepSchema>;
