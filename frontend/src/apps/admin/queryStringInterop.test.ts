import { describe, expect, it } from "vitest";
import { createRequire } from "node:module";

// `query-string@7` (pulled in by `ra-core`) is CommonJS and does
// `require("decode-uri-component")` at load time. Every `decode-uri-component`
// release >= 0.4 is ESM-only, so that `require` returns a module namespace
// object instead of the decoder function and query-string throws
// `TypeError: decodeComponent is not a function` on every `parse()` call.
// `ra-core` parses the location in `useListParams`, which is what powers the
// `page`, `perPage`, `sort` and `filter` URL parameters of every admin list.
//
// Anchor the require at query-string itself: decode-uri-component is only
// installed as a nested copy below `query-string/node_modules`, exactly the way
// query-string consumes it.
const queryStringRequire = createRequire(
  createRequire(import.meta.url).resolve("query-string"),
);

describe("query-string / decode-uri-component interop", () => {
  it("loads decode-uri-component as a callable function", () => {
    expect(typeof queryStringRequire("decode-uri-component")).toBe("function");
  });

  it("parses the list URL parameters react-admin reads", () => {
    const { parse } = queryStringRequire("query-string");

    expect(parse("?page=2&perPage=25&sort=id&order=DESC")).toEqual({
      page: "2",
      perPage: "25",
      sort: "id",
      order: "DESC",
    });
  });

  it("round-trips a JSON encoded filter value", () => {
    const { parse } = queryStringRequire("query-string");
    const filter = { q: "billed apparat" };

    const parsed = parse(
      `?filter=${encodeURIComponent(JSON.stringify(filter))}`,
    );

    expect(JSON.parse(String(parsed.filter))).toEqual(filter);
  });
});
