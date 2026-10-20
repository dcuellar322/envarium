interface HtmlNode {
  type?: string
  tagName?: string
  properties?: Record<string, unknown>
  children?: HtmlNode[]
}

/** Keep horizontally scrolling code and tables usable before scripts load. */
export function focusScrollableContent(node: HtmlNode): void {
  if (node.type === 'element' && (node.tagName === 'pre' || node.tagName === 'table')) {
    node.properties ??= {}
    node.properties.tabIndex ??= 0
  }
  node.children?.forEach(focusScrollableContent)
}

export default function rehypeKeyboardNavigation() {
  return focusScrollableContent
}
