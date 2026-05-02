import { z } from "zod";

export const slideSchema = z.object({
  id: z.number(),
  status: z.string().default("active"),
  origin_created_at: z.iso.datetime({ offset: true }).nullish(),
  created_at: z.iso.datetime({ offset: true }),
  content: z
    .object({
      type: z.string().default("slide"),
      title: z.string().optional(),
      body: z.string().optional(),
      media: z
        .object({
          local_url: z.string().optional(),
          mime_type: z.string(),
        })
        .nullish(),
    })
    .default({ type: "slide" }),
  author: z
    .object({
      display_name: z.string(),
      username: z.string().optional(),
      avatar: z
        .object({
          local_url: z.string().optional(),
          mime_type: z.string(),
        })
        .nullish(),
    })
    .nullish(),
  display_options: z.object({
    priority: z.number(),
    is_urgent: z.boolean(),
    allow_social_overlay: z.boolean(),
  }),
});

export type Slide = z.infer<typeof slideSchema>;
