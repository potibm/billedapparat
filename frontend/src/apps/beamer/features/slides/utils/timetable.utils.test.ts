/// <reference types="@testing-library/jest-dom" />
import { describe, it, expect } from "vitest";
import { extractTextFromNode, HastNode } from "./timetable.utils";

describe("extractTextFromNode", () => {
  describe("edge cases", () => {
    it("should return empty string for undefined node", () => {
      expect(extractTextFromNode(undefined)).toBe("");
    });

    it("should return empty string for node without value or children", () => {
      const node: HastNode = { type: "element", tagName: "div" };
      expect(extractTextFromNode(node)).toBe("");
    });

    it("should return empty string for node with empty children array", () => {
      const node: HastNode = { type: "element", tagName: "div", children: [] };
      expect(extractTextFromNode(node)).toBe("");
    });
  });

  describe("text nodes", () => {
    it("should extract text from a simple text node", () => {
      const node: HastNode = { type: "text", value: "Hello World" };
      expect(extractTextFromNode(node)).toBe("Hello World");
    });

    it("should return empty string for text node without value", () => {
      const node: HastNode = { type: "text" };
      expect(extractTextFromNode(node)).toBe("");
    });

    it("should preserve whitespace in text nodes", () => {
      const node: HastNode = { type: "text", value: "  spaces  " };
      expect(extractTextFromNode(node)).toBe("  spaces  ");
    });

    it("should handle newlines in text nodes", () => {
      const node: HastNode = { type: "text", value: "line1\nline2" };
      expect(extractTextFromNode(node)).toBe("line1\nline2");
    });
  });

  describe("nested children", () => {
    it("should extract text from single child", () => {
      const node: HastNode = {
        type: "element",
        tagName: "span",
        children: [{ type: "text", value: "Child text" }],
      };
      expect(extractTextFromNode(node)).toBe("Child text");
    });

    it("should concatenate text from multiple children", () => {
      const node: HastNode = {
        type: "element",
        tagName: "div",
        children: [
          { type: "text", value: "Hello " },
          { type: "text", value: "World" },
        ],
      };
      expect(extractTextFromNode(node)).toBe("Hello World");
    });

    it("should handle deeply nested structures", () => {
      const node: HastNode = {
        type: "element",
        tagName: "div",
        children: [
          {
            type: "element",
            tagName: "span",
            children: [
              {
                type: "element",
                tagName: "strong",
                children: [{ type: "text", value: "Deep" }],
              },
            ],
          },
        ],
      };
      expect(extractTextFromNode(node)).toBe("Deep");
    });

    it("should handle mixed element and text nodes", () => {
      const node: HastNode = {
        type: "element",
        tagName: "p",
        children: [
          { type: "text", value: "Before " },
          {
            type: "element",
            tagName: "em",
            children: [{ type: "text", value: "italic" }],
          },
          { type: "text", value: " after" },
        ],
      };
      expect(extractTextFromNode(node)).toBe("Before italic after");
    });

    it("should handle complex nested structure with multiple levels", () => {
      const node: HastNode = {
        type: "element",
        tagName: "table",
        children: [
          {
            type: "element",
            tagName: "thead",
            children: [
              {
                type: "element",
                tagName: "tr",
                children: [
                  {
                    type: "element",
                    tagName: "th",
                    children: [{ type: "text", value: "Header 1" }],
                  },
                  {
                    type: "element",
                    tagName: "th",
                    children: [{ type: "text", value: "Header 2" }],
                  },
                ],
              },
            ],
          },
        ],
      };
      expect(extractTextFromNode(node)).toBe("Header 1Header 2");
    });
  });

  describe("real-world HAST structures", () => {
    it("should extract text from a th element with text child", () => {
      const thNode: HastNode = {
        type: "element",
        tagName: "th",
        children: [{ type: "text", value: "Category Color" }],
      };
      expect(extractTextFromNode(thNode)).toBe("Category Color");
    });

    it("should extract text from td with multiple text nodes", () => {
      const tdNode: HastNode = {
        type: "element",
        tagName: "td",
        children: [
          { type: "text", value: "09:00" },
          { type: "text", value: " - " },
          { type: "text", value: "10:00" },
        ],
      };
      expect(extractTextFromNode(tdNode)).toBe("09:00 - 10:00");
    });

    it("should handle th with inline formatting", () => {
      const thNode: HastNode = {
        type: "element",
        tagName: "th",
        children: [
          { type: "text", value: "Start " },
          {
            type: "element",
            tagName: "strong",
            children: [{ type: "text", value: "Time" }],
          },
        ],
      };
      expect(extractTextFromNode(thNode)).toBe("Start Time");
    });
  });
});
