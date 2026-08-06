import { Slide } from "../../types/slide.schema";
import Markdown, { Components } from "react-markdown";
import remarkGfm from "remark-gfm";
import React, {
  HTMLAttributes,
  useMemo,
  ReactElement,
  createContext,
  use,
} from "react";
import { getContrastTextColor } from "@core/colors/colors";

interface HastNode {
  type: string;
  tagName?: string;
  value?: string;
  children?: HastNode[];
}

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

const TableHeaderContext = createContext<string[]>([]);

const extractTextFromNode = (node?: HastNode): string => {
  if (!node) return "";
  if (node.type === "text" && node.value) return node.value;
  if (node.children) return node.children.map(extractTextFromNode).join("");
  return "";
};

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

const markdownComponents: Components = {
  table: TimetableTable,

  tr: TimetableRow,

  td: ({ node: _node, className, ...props }) => (
    <td className={`p-3 align-middle ${className || ""}`.trim()} {...props} />
  ),
  th: ({ node: _node, className, ...props }) => (
    <th
      className={`p-3 text-left bg-gray-800 text-white font-bold ${className || ""}`.trim()}
      {...props}
    />
  ),
};

export const TimetableSlide = ({ slide }: { slide: Slide }) => {
  return (
    <div className="w-full h-full slide-timetable">
      <div className="flex items-center justify-center w-full h-full slide-timetable__container">
        {slide.content.body ? (
          <div className="text-gray-200 p-8 max-w-4xl mx-auto slide-timetable__content">
            <h1 className="slide-timetable__title">{slide.content.title}</h1>
            <div className="slide-timetable__body">
              <Markdown
                remarkPlugins={[remarkGfm]}
                components={markdownComponents}
              >
                {slide.content.body}
              </Markdown>
            </div>
          </div>
        ) : (
          <h2 className="text-5xl font-light text-gray-500 tracking-widest uppercase slide-timetable__fallback">
            Timetable: {slide.content.title}
          </h2>
        )}
      </div>
    </div>
  );
};
