import FilterAltIcon from "@mui/icons-material/FilterAlt";
import { FilterRulesCreate } from "./FilterRulesCreate";
import { FilterRulesList } from "./FilterRulesList";
import { FilterRulesEdit } from "./FilterRulesEdit";

export default {
  options: { label: "Filter Rules" },
  list: FilterRulesList,
  create: FilterRulesCreate,
  edit: FilterRulesEdit,
  icon: FilterAltIcon,
};
