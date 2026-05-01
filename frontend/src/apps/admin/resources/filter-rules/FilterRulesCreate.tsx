import { useMemo } from "react";
import {
  Create,
  SelectInput,
  SimpleForm,
  TextInput,
  useGetList,
} from "react-admin";

export const FilterRulesCreate = () => {
  const { data: sources, isLoading } = useGetList("sources", {
    pagination: { page: 1, perPage: 100 },
  });

  const sourceChoices = useMemo(() => {
    const defaultChoice = { id: "*", name: "All sources (*)" };
    if (!sources) return [defaultChoice];
    return [defaultChoice, ...sources];
  }, [sources]);

  const typeChoices = [
    { id: "language", name: "Language" },
    { id: "username", name: "Username" },
    { id: "display_name", name: "Display Name" },
  ];

  return (
    <Create title="Add Filter Rule">
      <SimpleForm defaultValues={{ source: "*", type: "username" }}>
        <SelectInput
          source="source"
          choices={sourceChoices}
          isLoading={isLoading}
          translateChoice={false}
          required
        />

        <SelectInput
          source="type"
          choices={typeChoices}
          translateChoice={false}
          required
        />

        <TextInput source="value" label="Value" fullWidth required />
      </SimpleForm>
    </Create>
  );
};

export default FilterRulesCreate;
