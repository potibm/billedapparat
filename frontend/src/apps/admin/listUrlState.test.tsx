import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import {
  AdminContext,
  Datagrid,
  List,
  Resource,
  TestMemoryRouter,
  TextField,
  type DataProvider,
  type RaRecord,
} from "react-admin";

const PER_PAGE = 10;
const TOTAL = 30;

const records: RaRecord[] = Array.from({ length: PER_PAGE }, (_, index) => ({
  id: index + 1,
  title: `News ${index + 1}`,
}));

const NewsList = () => (
  <List perPage={PER_PAGE} sort={{ field: "title", order: "ASC" }}>
    <Datagrid>
      <TextField source="id" />
      <TextField source="title" />
    </Datagrid>
  </List>
);

const createDataProvider = () =>
  ({
    getList: vi.fn().mockResolvedValue({ data: records, total: TOTAL }),
    getOne: vi.fn().mockResolvedValue({ data: records[0] }),
    getMany: vi.fn().mockResolvedValue({ data: records }),
    getManyReference: vi
      .fn()
      .mockResolvedValue({ data: records, total: TOTAL }),
  }) as unknown as DataProvider;

const renderNewsList = (search: string) => {
  const dataProvider = createDataProvider();
  let locationSearch = search;

  render(
    <TestMemoryRouter
      initialEntries={[`/news${search}`]}
      locationCallback={(location) => {
        locationSearch = location.search;
      }}
    >
      <AdminContext dataProvider={dataProvider}>
        <Resource name="news" list={NewsList} />
      </AdminContext>
    </TestMemoryRouter>,
  );

  const getList = dataProvider.getList as unknown as ReturnType<typeof vi.fn>;

  return { getList, getSearch: () => locationSearch };
};

describe("admin list URL state", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("applies page, perPage, sort and filter from the URL", async () => {
    const { getList } = renderNewsList(
      `?page=2&perPage=${PER_PAGE}&sort=id&order=DESC` +
        `&filter=${encodeURIComponent(JSON.stringify({ q: "billed apparat" }))}`,
    );

    await waitFor(() => expect(getList).toHaveBeenCalled());

    expect(getList).toHaveBeenLastCalledWith(
      "news",
      expect.objectContaining({
        pagination: { page: 2, perPage: PER_PAGE },
        sort: { field: "id", order: "DESC" },
        filter: { q: "billed apparat" },
      }),
    );
  });

  it("keeps the filter while paginating and syncs the page into the URL", async () => {
    const { getList, getSearch } = renderNewsList(
      `?filter=${encodeURIComponent(JSON.stringify({ q: "billed apparat" }))}`,
    );

    await screen.findByText("News 1");

    const paginator = await screen.findByRole("navigation", {
      name: "pagination navigation",
    });
    await userEvent.click(within(paginator).getByText("2"));

    await waitFor(() => expect(getSearch()).toContain("page=2"));

    expect(getSearch()).toContain(
      `filter=${encodeURIComponent(JSON.stringify({ q: "billed apparat" }))}`,
    );
    expect(getList).toHaveBeenLastCalledWith(
      "news",
      expect.objectContaining({
        pagination: { page: 2, perPage: PER_PAGE },
        filter: { q: "billed apparat" },
      }),
    );
  });
});
