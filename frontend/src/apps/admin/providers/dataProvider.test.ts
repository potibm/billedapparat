import { describe, it, expect, vi, beforeEach, type Mock } from "vitest";

const mockGetAccessToken = vi.fn();
vi.mock("./authProvider", () => ({
  getAccessToken: mockGetAccessToken,
}));

const mockBaseProvider = {
  getList: vi.fn(),
  getOne: vi.fn(),
  create: vi.fn(),
  update: vi.fn(),
  delete: vi.fn(),
  deleteMany: vi.fn(),
};

vi.mock("ra-data-json-server", () => ({
  default: vi.fn(() => mockBaseProvider),
}));

vi.mock("react-admin", async () => {
  const actual =
    await vi.importActual<typeof import("react-admin")>("react-admin");
  return {
    ...actual,
    fetchUtils: {
      fetchJson: vi.fn(),
    },
  };
});

const { dataProvider } = await import("./dataProvider");

describe("dataProvider", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockGetAccessToken.mockResolvedValue(null);
  });

  describe("getList", () => {
    it("should add type filter for mapped slide resources", async () => {
      const params = { filter: { status: "active" }, pagination: {}, sort: {} };
      mockBaseProvider.getList.mockResolvedValue({ data: [], total: 0 });

      await dataProvider.getList("sponsor-slides", params as never);

      expect(mockBaseProvider.getList).toHaveBeenCalledWith("slides", {
        ...params,
        filter: { status: "active", type: "sponsor" },
      });
    });

    it("should map news-slides to type news", async () => {
      const params = { filter: {}, pagination: {}, sort: {} };
      mockBaseProvider.getList.mockResolvedValue({ data: [], total: 0 });

      await dataProvider.getList("news-slides", params as never);

      expect(mockBaseProvider.getList).toHaveBeenCalledWith("slides", {
        ...params,
        filter: { type: "news" },
      });
    });

    it("should map timetable-slides to type timetable", async () => {
      const params = { filter: {}, pagination: {}, sort: {} };
      mockBaseProvider.getList.mockResolvedValue({ data: [], total: 0 });

      await dataProvider.getList("timetable-slides", params as never);

      expect(mockBaseProvider.getList).toHaveBeenCalledWith("slides", {
        ...params,
        filter: { type: "timetable" },
      });
    });

    it("should map social-medias-slides to type social.media", async () => {
      const params = { filter: {}, pagination: {}, sort: {} };
      mockBaseProvider.getList.mockResolvedValue({ data: [], total: 0 });

      await dataProvider.getList("social-medias-slides", params as never);

      expect(mockBaseProvider.getList).toHaveBeenCalledWith("slides", {
        ...params,
        filter: { type: "social.media" },
      });
    });

    it("should map social-text-slides to type social.text", async () => {
      const params = { filter: {}, pagination: {}, sort: {} };
      mockBaseProvider.getList.mockResolvedValue({ data: [], total: 0 });

      await dataProvider.getList("social-text-slides", params as never);

      expect(mockBaseProvider.getList).toHaveBeenCalledWith("slides", {
        ...params,
        filter: { type: "social.text" },
      });
    });

    it("should pass through unmapped resources directly", async () => {
      const params = { filter: {}, pagination: {}, sort: {} };
      mockBaseProvider.getList.mockResolvedValue({ data: [], total: 0 });

      await dataProvider.getList("users", params as never);

      expect(mockBaseProvider.getList).toHaveBeenCalledWith("users", params);
    });
  });

  describe("getOne", () => {
    it("should map slide resources to slides for getOne", async () => {
      mockBaseProvider.getOne.mockResolvedValue({ data: { id: 1 } });

      await dataProvider.getOne("sponsor-slides", { id: 1 });

      expect(mockBaseProvider.getOne).toHaveBeenCalledWith("slides", { id: 1 });
    });

    it("should pass through unmapped resources for getOne", async () => {
      mockBaseProvider.getOne.mockResolvedValue({ data: { id: 1 } });

      await dataProvider.getOne("users", { id: 1 });

      expect(mockBaseProvider.getOne).toHaveBeenCalledWith("users", { id: 1 });
    });
  });

  describe("delete", () => {
    it("should map slide resources to slides for delete", async () => {
      mockBaseProvider.delete.mockResolvedValue({ data: { id: 1 } });

      await dataProvider.delete("news-slides", {
        id: 1,
        previousData: { id: 1 },
      });

      expect(mockBaseProvider.delete).toHaveBeenCalledWith("slides", {
        id: 1,
        previousData: { id: 1 },
      });
    });

    it("should pass through unmapped resources for delete", async () => {
      mockBaseProvider.delete.mockResolvedValue({ data: { id: 1 } });

      await dataProvider.delete("users", { id: 1, previousData: { id: 1 } });

      expect(mockBaseProvider.delete).toHaveBeenCalledWith("users", {
        id: 1,
        previousData: { id: 1 },
      });
    });
  });

  describe("deleteMany", () => {
    it("should map slide resources to slides for deleteMany", async () => {
      mockBaseProvider.deleteMany.mockResolvedValue({ data: [1, 2] });

      await dataProvider.deleteMany("sponsor-slides", { ids: [1, 2] });

      expect(mockBaseProvider.deleteMany).toHaveBeenCalledWith("slides", {
        ids: [1, 2],
      });
    });

    it("should pass through unmapped resources for deleteMany", async () => {
      mockBaseProvider.deleteMany.mockResolvedValue({ data: [1] });

      await dataProvider.deleteMany("users", { ids: [1] });

      expect(mockBaseProvider.deleteMany).toHaveBeenCalledWith("users", {
        ids: [1],
      });
    });
  });

  describe("create", () => {
    it("should delegate to baseProvider.create for non-slide resources", async () => {
      mockBaseProvider.create.mockResolvedValue({ data: { id: 1 } });

      await dataProvider.create("users", { data: { name: "test" } });

      expect(mockBaseProvider.create).toHaveBeenCalledWith("users", {
        data: { name: "test" },
      });
    });

    it("should call baseProvider.create for slide resources without file upload", async () => {
      mockBaseProvider.create.mockResolvedValue({ data: { id: 1 } });

      await dataProvider.create("sponsor-slides", {
        data: { status: "active", content: { title: "Test" } },
      });

      expect(mockBaseProvider.create).toHaveBeenCalledWith(
        "slides",
        expect.objectContaining({
          data: expect.objectContaining({
            status: "active",
            content: expect.objectContaining({ type: "sponsor" }),
          }),
        }),
      );
    });

    it("should use FormData for slide resources with file upload", async () => {
      const file = new File(["content"], "test.png", { type: "image/png" });
      mockGetAccessToken.mockResolvedValue("test-token");

      const mockResponse = { id: 1, status: "active" };
      globalThis.fetch = vi.fn().mockResolvedValue({
        ok: true,
        json: () => Promise.resolve(mockResponse),
      });

      const result = await dataProvider.create("sponsor-slides", {
        data: {
          status: "active",
          content: { title: "My Slide", body: "Body text" },
          author: { display_name: "Author" },
          display_options: { priority: 2 },
          media_upload: { rawFile: file },
        },
      });

      expect(globalThis.fetch).toHaveBeenCalledWith(
        "/api/admin/slides",
        expect.objectContaining({
          method: "POST",
          body: expect.any(FormData),
        }),
      );

      const fetchCall = (globalThis.fetch as Mock).mock.calls[0];
      const formData = fetchCall[1].body as FormData;
      expect(formData.get("status")).toBe("active");
      expect(formData.get("content.type")).toBe("sponsor");
      expect(formData.get("content.title")).toBe("My Slide");
      expect(formData.get("content.body")).toBe("Body text");
      expect(formData.get("author.display_name")).toBe("Author");
      expect(formData.get("display_options.priority")).toBe("2");
      expect(formData.get("media_upload")).toBe(file);

      expect(result).toEqual({ data: mockResponse });
    });

    it("should set Authorization header on file upload when token exists", async () => {
      const file = new File(["content"], "test.png", { type: "image/png" });
      mockGetAccessToken.mockResolvedValue("my-token");

      globalThis.fetch = vi.fn().mockResolvedValue({
        ok: true,
        json: () => Promise.resolve({ id: 1 }),
      });

      await dataProvider.create("sponsor-slides", {
        data: { media_upload: { rawFile: file } },
      });

      const fetchCall = (globalThis.fetch as Mock).mock.calls[0];
      const headers = fetchCall[1].headers as Headers;
      expect(headers.get("Authorization")).toBe("Bearer my-token");
    });

    it("should not set Authorization header when token is null", async () => {
      const file = new File(["content"], "test.png", { type: "image/png" });
      mockGetAccessToken.mockResolvedValue(null);

      globalThis.fetch = vi.fn().mockResolvedValue({
        ok: true,
        json: () => Promise.resolve({ id: 1 }),
      });

      await dataProvider.create("sponsor-slides", {
        data: { media_upload: { rawFile: file } },
      });

      const fetchCall = (globalThis.fetch as Mock).mock.calls[0];
      const headers = fetchCall[1].headers as Headers;
      expect(headers.get("Authorization")).toBeNull();
    });

    it("should use default values for missing fields in file upload", async () => {
      const file = new File(["content"], "test.png", { type: "image/png" });
      mockGetAccessToken.mockResolvedValue(null);

      globalThis.fetch = vi.fn().mockResolvedValue({
        ok: true,
        json: () => Promise.resolve({ id: 1 }),
      });

      await dataProvider.create("sponsor-slides", {
        data: { media_upload: { rawFile: file } },
      });

      const fetchCall = (globalThis.fetch as Mock).mock.calls[0];
      const formData = fetchCall[1].body as FormData;
      expect(formData.get("status")).toBe("active");
      expect(formData.get("content.type")).toBe("sponsor");
      expect(formData.get("content.title")).toBe("");
      expect(formData.get("content.body")).toBe("");
      expect(formData.get("author.display_name")).toBe("");
      expect(formData.get("display_options.priority")).toBe("1");
    });
  });

  describe("update", () => {
    it("should delegate to baseProvider.update for non-slide resources", async () => {
      mockBaseProvider.update.mockResolvedValue({ data: { id: 1 } });

      await dataProvider.update("users", {
        id: 1,
        data: { name: "test" },
        previousData: {},
      });

      expect(mockBaseProvider.update).toHaveBeenCalledWith("users", {
        id: 1,
        data: { name: "test" },
        previousData: {},
      });
    });

    it("should call baseProvider.update for slide resources without file upload", async () => {
      mockBaseProvider.update.mockResolvedValue({ data: { id: 1 } });

      await dataProvider.update("news-slides", {
        id: 1,
        data: { status: "active" },
        previousData: {},
      });

      expect(mockBaseProvider.update).toHaveBeenCalledWith(
        "slides",
        expect.objectContaining({
          id: 1,
          data: expect.objectContaining({
            content: expect.objectContaining({ type: "news" }),
          }),
        }),
      );
    });

    it("should use FormData with PUT for slide resources with file upload", async () => {
      const file = new File(["content"], "test.png", { type: "image/png" });
      mockGetAccessToken.mockResolvedValue("test-token");

      globalThis.fetch = vi.fn().mockResolvedValue({
        ok: true,
        json: () => Promise.resolve({ id: 5 }),
      });

      const result = await dataProvider.update("sponsor-slides", {
        id: 5,
        data: {
          status: "inactive",
          content: { title: "Updated" },
          media_upload: { rawFile: file },
        },
        previousData: {},
      });

      expect(globalThis.fetch).toHaveBeenCalledWith(
        "/api/admin/slides/5",
        expect.objectContaining({
          method: "PUT",
          body: expect.any(FormData),
        }),
      );

      expect(result).toEqual({ data: { id: 5 } });
    });

    it("should throw on failed file upload response", async () => {
      const file = new File(["content"], "test.png", { type: "image/png" });
      mockGetAccessToken.mockResolvedValue(null);

      globalThis.fetch = vi.fn().mockResolvedValue({
        ok: false,
        statusText: "Internal Server Error",
        text: () => Promise.resolve("Something went wrong"),
      });

      await expect(
        dataProvider.create("sponsor-slides", {
          data: { media_upload: { rawFile: file } },
        }),
      ).rejects.toThrow("Something went wrong");
    });

    it("should fall back to statusText on failed response with empty body", async () => {
      const file = new File(["content"], "test.png", { type: "image/png" });
      mockGetAccessToken.mockResolvedValue(null);

      globalThis.fetch = vi.fn().mockResolvedValue({
        ok: false,
        statusText: "Bad Gateway",
        text: () => Promise.resolve(""),
      });

      await expect(
        dataProvider.create("sponsor-slides", {
          data: { media_upload: { rawFile: file } },
        }),
      ).rejects.toThrow("Bad Gateway");
    });
  });
});
