import { z } from "zod";

export const slideSchema = z.object({
  id: z.number(),
  status: z.string().default("active"),
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
      username: z.string().optional(),
      displayname: z.string().optional(),
      media: z
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
  }),
});

export type Slide = z.infer<typeof slideSchema>;
