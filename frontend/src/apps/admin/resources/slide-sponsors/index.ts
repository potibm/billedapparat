import { SponsorCreate } from "./SponsorCreate";
import { SponsorList } from "./SponsorList";
import { SponsorEdit } from "./SponsorEdit";
import BusinessIcon from "@mui/icons-material/Business";

export default {
  options: { label: "Sponsor Slides" },
  list: SponsorList,
  create: SponsorCreate,
  edit: SponsorEdit,
  icon: BusinessIcon,
};
