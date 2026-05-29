import {
  useRecordContext,
  useResourceContext,
  useUpdate,
  useNotify,
  FieldProps,
} from "react-admin";
import { Switch } from "@mui/material";

export interface StatusToggleFieldProps extends FieldProps {
  source: string;
  activeValue?: string | number | boolean;
  inactiveValue?: string | number | boolean;
}

export const StatusToggleField = ({
  source,
  activeValue = "active",
  inactiveValue = "inactive",
  ...props
}: StatusToggleFieldProps) => {
  const record = useRecordContext<Record<string, unknown>>();
  const resource = useResourceContext();
  const notify = useNotify();

  const [update, { isLoading }] = useUpdate();

  if (!record) return null;

  const isChecked = record[source] === activeValue;

  const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    event.stopPropagation();

    const newValue = event.target.checked ? activeValue : inactiveValue;

    update(
      resource,
      {
        id: record.id as string | number,
        data: { ...record, [source]: newValue },
        previousData: record,
      },
      {
        mutationMode: "optimistic",
        onSuccess: () => {
          notify(`Status updated`, { type: "success" });
        },
        onError: (error) => {
          const errorMessage =
            error instanceof Error ? error.message : String(error);

          notify(`Error while updating status: ${errorMessage}`, {
            type: "error",
          });
        },
      },
    );
  };

  return (
    <Switch
      checked={isChecked}
      onChange={handleChange}
      disabled={isLoading}
      onClick={(e) => e.stopPropagation()}
      {...props}
    />
  );
};

StatusToggleField.defaultProps = {
  addLabel: true,
};
