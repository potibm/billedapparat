import { Create, SimpleForm } from "react-admin";
import { FilterRuleInputs } from "./FilterRuleInputs";

export const FilterRulesCreate = () => {
  return (
    <Create title="Add Filter Rule">
      <SimpleForm defaultValues={{ source: "*", type: "username" }}>
        <FilterRuleInputs />
      </SimpleForm>
    </Create>
  );
};

export default FilterRulesCreate;
