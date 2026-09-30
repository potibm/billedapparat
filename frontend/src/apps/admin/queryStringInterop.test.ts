import { describe, expect, it } from "vitest";
import { parse, stringify } from "query-string";

// `query-string@7` (pulled in by `ra-core`) is CommonJS and does
// `require("decode-uri-component")` at load time. Every `decode-uri-component`
// release >= 0.4 is ESM-only, so that `require` hands back a module namespace
// object instead of the decoder function and query-string throws
// `TypeError: decodeComponent is not a function` on every `parse()` call.
//
// `ra-core` parses the location in `useListParams`, which backs the `page`,
// `perPage`, `sort` and `filter` URL parameters of every admin list. Do not
// pin `decode-uri-component` to >= 0.4 in `frontend/package.json` to silence
// GHSA-vcc3-ghjq-m6fr - the tests below fail if you do.
describe("query-string list URL parsing", () => {
  it("parses the list URL parameters react-admin reads", () => {
    expect(parse("?page=2&perPage=25&sort=id&order=DESC")).toEqual({
      page: "2",
      perPage: "25",
      sort: "id",
      order: "DESC",
    });
  });

  it("round-trips a JSON encoded filter value", () => {
    const filter = { q: "billed apparat" };

    const parsed = parse(
      `?filter=${encodeURIComponent(JSON.stringify(filter))}`,
    );

    expect(JSON.parse(String(parsed.filter))).toEqual(filter);
  });

  it("serializes filter and page back into the query string", () => {
    expect(
      stringify({
        filter: JSON.stringify({ q: "billed apparat" }),
        page: 2,
        perPage: 25,
      }),
    ).toContain("page=2");
  });
});
