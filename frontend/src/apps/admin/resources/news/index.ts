import { NewsEntityList } from "./NewsEntityList";
import ArticleIcon from "@mui/icons-material/Article";
import { NewsEntityShow } from "./NewsEntityShow";

export default {
  options: { label: "News" },
  list: NewsEntityList,
  show: NewsEntityShow,
  icon: ArticleIcon,
};
