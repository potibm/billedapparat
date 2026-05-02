import { Edit, SimpleForm } from "react-admin";
import { FilterRuleInputs } from "./FilterRuleInputs";

export const FilterRulesEdit = () => {
  return (
    <Edit title="Edit Filter Rule">
      <SimpleForm defaultValues={{ source: "*", type: "username" }}>
        <FilterRuleInputs />
      </SimpleForm>
    </Edit>
  );
};

export default FilterRulesEdit;
