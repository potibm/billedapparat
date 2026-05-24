import { ImagePreviewField } from "@admin/components/fields/ImagePreviewField";
import { BaseSocialShow } from "@admin/components/show/BaseSocialShow";
import { ShowProps } from "react-admin";

export const SocialMediaShow = (props: ShowProps) => (
  <BaseSocialShow {...props} title="View Social Media">
    <ImagePreviewField
      source="content.media.local_url"
      label="Media"
      maxWidth={480}
      maxHeight={270}
    />
  </BaseSocialShow>
);
