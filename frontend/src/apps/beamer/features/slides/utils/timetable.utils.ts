export interface HastNode {
  type: string;
  tagName?: string;
  value?: string;
  children?: HastNode[];
}

export const extractTextFromNode = (node?: HastNode): string => {
  if (!node) return "";
  if (node.type === "text" && node.value) return node.value;
  if (node.children) return node.children.map(extractTextFromNode).join("");
  return "";
};
