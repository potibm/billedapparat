import React, {
  useMemo,
  ReactElement,
  createContext,
  use,
  HTMLAttributes,
  TdHTMLAttributes,
  ThHTMLAttributes,
} from "react";
import { getContrastTextColor } from "@core/colors/colors";
import { HastNode, extractTextFromNode } from "../../utils/timetable.utils";

interface MarkdownCellProps {
  node?: HastNode;
  children?: React.ReactNode;
  className?: string;
}
interface TableComponentProps extends HTMLAttributes<HTMLTableElement> {
  node?: HastNode;
}

interface TableRowComponentProps extends HTMLAttributes<HTMLTableRowElement> {
  node?: HastNode;
}

interface TdComponentProps extends TdHTMLAttributes<HTMLTableCellElement> {
  node?: HastNode;
}

interface ThComponentProps extends ThHTMLAttributes<HTMLTableCellElement> {
  node?: HastNode;
}

const TableHeaderContext = createContext<string[]>([]);

export const TimetableTable = ({ node, ...props }: TableComponentProps) => {
  const headers = useMemo(() => {
    let extractedHeaders: string[] = [];

    if (node?.type === "element") {
      const thead = (node as HastNode).children?.find(
        (n) => n.tagName === "thead",
      );
      const tr = thead?.children?.find((n) => n.tagName === "tr");

      if (tr?.children) {
        const ths = tr.children.filter((n) => n.tagName === "th");
        extractedHeaders = ths.map((th) => extractTextFromNode(th).trim());
      }
    }

    return extractedHeaders;
  }, [node]);

  return (
    <TableHeaderContext value={headers}>
      <table className="slide-timetable__table" {...props} />
    </TableHeaderContext>
  );
};

export const TimetableRow = ({
  node: _node,
  children,
  ...props
}: TableRowComponentProps) => {
  const headers = use(TableHeaderContext);

  const { colorColumnIndex, categoryColumnIndex } = useMemo(() => {
    const colorIdx = headers.findIndex(
      (h) => h.toLowerCase() === "category color",
    );
    const categoryIdx = headers.findIndex(
      (h) => h.toLowerCase() === "category",
    );
    return {
      colorColumnIndex: colorIdx >= 0 ? colorIdx : -1,
      categoryColumnIndex: categoryIdx >= 0 ? categoryIdx : -1,
    };
  }, [headers]);

  // eslint-disable-next-line @eslint-react/no-children-to-array
  const cells = React.Children.toArray(
    children,
  ) as ReactElement<MarkdownCellProps>[];

  const colorCell = colorColumnIndex >= 0 ? cells[colorColumnIndex] : undefined;
  let hexColor = "#9ca3af";
  if (colorCell?.props?.children) {
    const content = colorCell.props.children;
    const colorValue = Array.isArray(content) ? content[0] : content;
    if (typeof colorValue === "string" && colorValue.trim()) {
      const trimmedColor = colorValue.trim();
      if (/^#[0-9a-fA-F]{6}$/.test(trimmedColor)) {
        hexColor = trimmedColor;
      }
    }
  }

  const modifiedCells = cells
    .map((cell, index) => {
      const tagName = cell.props?.node?.tagName;

      if (tagName === "td" || tagName === "th") {
        const headerName = headers[index] || `col-${index}`;
        const formattedName = headerName.toLowerCase().replace(/\s+/g, "-");
        const cellClass = `slide-timetable__cell-${formattedName}`;

        let cellContent = cell.props.children;

        if (index === categoryColumnIndex && tagName === "td") {
          const textColorClass = getContrastTextColor(hexColor);
          cellContent = (
            <span
              style={{ backgroundColor: hexColor }}
              className={`px-3 py-1 rounded-full text-xs font-bold uppercase tracking-wide whitespace-nowrap inline-block ${textColorClass}`}
            >
              {cellContent}
            </span>
          );
        }

        // eslint-disable-next-line @eslint-react/no-clone-element
        return React.cloneElement(cell, {
          className: cellClass,
          children: cellContent,
        });
      }

      return cell;
    })
    .filter((_, index) => index !== colorColumnIndex);

  return (
    <tr className="border-b" {...props}>
      {modifiedCells}
    </tr>
  );
};

export const TimetableTd = ({
  node: _node,
  className,
  ...props
}: TdComponentProps) => (
  <td className={`p-3 align-middle ${className || ""}`.trim()} {...props} />
);

export const TimetableTh = ({
  node: _node,
  className,
  ...props
}: ThComponentProps) => (
  <th
    className={`p-3 text-left bg-gray-800 text-white font-bold ${className || ""}`.trim()}
    {...props}
  />
);
