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
  // eslint-disable-next-line @eslint-react/no-children-to-array
  const cells = React.Children.toArray(
    children,
  ) as ReactElement<MarkdownCellProps>[];

  const colorCell = cells[3];
  let hexColor = "#9ca3af";
  if (colorCell?.props?.children) {
    const content = colorCell.props.children;
    hexColor = Array.isArray(content) ? String(content[0]) : String(content);
  }

  const modifiedCells = cells
    .map((cell, index) => {
      const tagName = cell.props?.node?.tagName;

      // Now we intervene for both td AND th
      if (tagName === "td" || tagName === "th") {
        const headerName = headers[index] || `col-${index}`;
        const formattedName = headerName.toLowerCase().replace(/\s+/g, "-");
        const cellClass = `slide-timetable__cell-${formattedName}`;

        let cellContent = cell.props.children;

        // Pill styling only applies to the category td (index 2)
        if (index === 2 && tagName === "td") {
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

        // Modify cell: pass class and (if changed) children
        // eslint-disable-next-line @eslint-react/no-clone-element
        return React.cloneElement(cell, {
          className: cellClass, // Attach class as prop to the cell
          children: cellContent,
        });
      }

      return cell;
    })
    .filter((_, index) => index !== 3);

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
