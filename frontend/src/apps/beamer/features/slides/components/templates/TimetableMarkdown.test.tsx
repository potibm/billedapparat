/// <reference types="@testing-library/jest-dom" />
import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import Markdown from "react-markdown";
import remarkGfm from "remark-gfm";
import {
  TimetableTable,
  TimetableRow,
  TimetableTd,
  TimetableTh,
} from "./TimetableMarkdown";

const markdownComponents = {
  table: TimetableTable,
  tr: TimetableRow,
  td: TimetableTd,
  th: TimetableTh,
};

const renderMarkdown = (markdown: string) => {
  return render(
    <Markdown remarkPlugins={[remarkGfm]} components={markdownComponents}>
      {markdown}
    </Markdown>,
  );
};

describe("TimetableTable", () => {
  it("should extract headers from markdown table and render with context", () => {
    const markdown = `| Start | End | Category | Category Color |
| --- | --- | --- | --- |
| 09:00 | 10:00 | Workshop | #ff0000 |`;

    renderMarkdown(markdown);

    const table = screen.getByRole("table");
    expect(table).toBeInTheDocument();
    expect(table).toHaveClass("slide-timetable__table");
  });

  it("should render table headers correctly", () => {
    const markdown = `| Start | End | Category |
| --- | --- | --- |
| 09:00 | 10:00 | Workshop |`;

    renderMarkdown(markdown);

    const headers = screen.getAllByRole("columnheader");
    expect(headers).toHaveLength(3);
    expect(headers[0]).toHaveTextContent("Start");
    expect(headers[1]).toHaveTextContent("End");
    expect(headers[2]).toHaveTextContent("Category");
  });

  it("should render table cells correctly", () => {
    const markdown = `| Start | End |
| --- | --- |
| 09:00 | 10:00 |`;

    renderMarkdown(markdown);

    const cells = screen.getAllByRole("cell");
    expect(cells).toHaveLength(2);
    expect(cells[0]).toHaveTextContent("09:00");
    expect(cells[1]).toHaveTextContent("10:00");
  });
});

describe("TimetableRow", () => {
  it("should apply dynamic class names based on header names", () => {
    const markdown = `| Start | End | Category | Category Color |
| --- | --- | --- | --- |
| 09:00 | 10:00 | Workshop | #ff0000 |`;

    renderMarkdown(markdown);

    const cells = screen.getAllByRole("cell");
    expect(cells[0]).toHaveClass("slide-timetable__cell-start");
    expect(cells[1]).toHaveClass("slide-timetable__cell-end");
    expect(cells[2]).toHaveClass("slide-timetable__cell-category");
  });

  it("should wrap category cell in a pill with background color", () => {
    const markdown = `| Start | End | Category | Category Color |
| --- | --- | --- | --- |
| 09:00 | 10:00 | Workshop | #ff0000 |`;

    renderMarkdown(markdown);

    const cells = screen.getAllByRole("cell");
    const categoryCell = cells[2];
    const pill = categoryCell.querySelector("span");

    expect(pill).toBeInTheDocument();
    expect(pill).toHaveStyle({ backgroundColor: "#ff0000" });
    expect(pill).toHaveClass("px-3", "py-1", "rounded-full");
    expect(pill).toHaveTextContent("Workshop");
  });

  it("should remove the color column from final output", () => {
    const markdown = `| Start | End | Category | Category Color |
| --- | --- | --- | --- |
| 09:00 | 10:00 | Workshop | #ff0000 |`;

    renderMarkdown(markdown);

    const cells = screen.getAllByRole("cell");
    expect(cells).toHaveLength(3);
    expect(cells[0]).toHaveTextContent("09:00");
    expect(cells[1]).toHaveTextContent("10:00");
    expect(cells[2]).toHaveTextContent("Workshop");
  });

  it("should use default gray color when color cell is empty", () => {
    const markdown = `| Start | End | Category | Category Color |
| --- | --- | --- | --- |
| 09:00 | 10:00 | Workshop |  |`;

    renderMarkdown(markdown);

    const cells = screen.getAllByRole("cell");
    const categoryCell = cells[2];
    const pill = categoryCell.querySelector("span");

    expect(pill).toHaveStyle({ backgroundColor: "#9ca3af" });
  });

  it("should use default gray color when color format is invalid", () => {
    const markdown = `| Start | End | Category | Category Color |
| --- | --- | --- | --- |
| 09:00 | 10:00 | Workshop | invalid-color |`;

    renderMarkdown(markdown);

    const cells = screen.getAllByRole("cell");
    const categoryCell = cells[2];
    const pill = categoryCell.querySelector("span");

    expect(pill).toHaveStyle({ backgroundColor: "#9ca3af" });
  });

  it("should handle missing category color column", () => {
    const markdown = `| Start | End | Category |
| --- | --- | --- |
| 09:00 | 10:00 | Workshop |`;

    renderMarkdown(markdown);

    const cells = screen.getAllByRole("cell");
    expect(cells).toHaveLength(3);

    const categoryCell = cells[2];
    const pill = categoryCell.querySelector("span");
    expect(pill).toHaveStyle({ backgroundColor: "#9ca3af" });
  });

  it("should handle missing category column", () => {
    const markdown = `| Start | End | Category Color |
| --- | --- | --- |
| 09:00 | 10:00 | #ff0000 |`;

    renderMarkdown(markdown);

    const cells = screen.getAllByRole("cell");
    expect(cells).toHaveLength(2);
    expect(cells[0]).toHaveTextContent("09:00");
    expect(cells[1]).toHaveTextContent("10:00");
  });

  it("should apply border-b class to row", () => {
    const markdown = `| Start |
| --- |
| 09:00 |`;

    renderMarkdown(markdown);

    const rows = screen.getAllByRole("row");
    expect(rows).toHaveLength(2);
    expect(rows[0]).toHaveClass("border-b");
    expect(rows[1]).toHaveClass("border-b");
  });

  it("should handle headers with spaces in names", () => {
    const markdown = `| Start Time | End Time | Event Category |
| --- | --- | --- |
| 09:00 | 10:00 | Workshop |`;

    renderMarkdown(markdown);

    const cells = screen.getAllByRole("cell");
    expect(cells[0]).toHaveClass("slide-timetable__cell-start-time");
    expect(cells[1]).toHaveClass("slide-timetable__cell-end-time");
    expect(cells[2]).toHaveClass("slide-timetable__cell-event-category");
  });

  it("should handle multiple data rows", () => {
    const markdown = `| Start | End | Category | Category Color |
| --- | --- | --- | --- |
| 09:00 | 10:00 | Workshop | #ff0000 |
| 10:00 | 11:00 | Lecture | #00ff00 |`;

    renderMarkdown(markdown);

    const rows = screen.getAllByRole("row");
    expect(rows).toHaveLength(3);

    const cells = screen.getAllByRole("cell");
    expect(cells).toHaveLength(6);

    const firstCategoryPill = cells[2].querySelector("span");
    expect(firstCategoryPill).toHaveStyle({ backgroundColor: "#ff0000" });

    const secondCategoryPill = cells[5].querySelector("span");
    expect(secondCategoryPill).toHaveStyle({ backgroundColor: "#00ff00" });
  });
});

describe("TimetableTd", () => {
  it("should render children correctly", () => {
    const markdown = `| Content |
| --- |
| Test Content |`;

    renderMarkdown(markdown);

    const cell = screen.getByRole("cell");
    expect(cell).toHaveTextContent("Test Content");
  });

  it("should apply default classes", () => {
    const markdown = `| Content |
| --- |
| Test |`;

    renderMarkdown(markdown);

    const cell = screen.getByRole("cell");
    expect(cell).toHaveClass("p-3", "align-middle");
  });

  it("should merge dynamic className with default classes", () => {
    const markdown = `| Start |
| --- |
| 09:00 |`;

    renderMarkdown(markdown);

    const cell = screen.getByRole("cell");
    expect(cell).toHaveClass(
      "p-3",
      "align-middle",
      "slide-timetable__cell-start",
    );
  });
});

describe("TimetableTh", () => {
  it("should render children correctly", () => {
    const markdown = `| Header |
| --- |
| Content |`;

    renderMarkdown(markdown);

    const header = screen.getAllByRole("columnheader")[0];
    expect(header).toHaveTextContent("Header");
  });

  it("should apply default classes", () => {
    const markdown = `| Header |
| --- |
| Content |`;

    renderMarkdown(markdown);

    const header = screen.getAllByRole("columnheader")[0];
    expect(header).toHaveClass(
      "p-3",
      "text-left",
      "bg-gray-800",
      "text-white",
      "font-bold",
    );
  });

  it("should merge dynamic className with default classes", () => {
    const markdown = `| Start Time |
| --- |
| 09:00 |`;

    renderMarkdown(markdown);

    const header = screen.getAllByRole("columnheader")[0];
    expect(header).toHaveClass(
      "p-3",
      "text-left",
      "bg-gray-800",
      "text-white",
      "font-bold",
      "slide-timetable__cell-start-time",
    );
  });
});
