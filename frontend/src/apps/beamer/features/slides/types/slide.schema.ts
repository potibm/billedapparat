import { z } from "zod";

export const slideSchema = z.object({
  id: z.number(),
  status: z.string().default("active"),
  content: z
    .object({
      type: z.string().default("slide"),
      text: z.string().optional(),
    })
    .default({ type: "slide" }),
  media_url_original: z.string().optional(),
  display_options: z
    .object({
      priority: z.number(),
      is_urgent: z.boolean(),
    })
    .optional(),
});

export type Slide = z.infer<typeof slideSchema>;
