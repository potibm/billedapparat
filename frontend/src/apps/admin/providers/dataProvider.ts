// src/apps/admin/dataProvider.ts
import jsonServerProvider from "ra-data-json-server";
import {
  CreateParams,
  DataProvider,
  UpdateParams,
  fetchUtils,
} from "react-admin";
import { getAccessToken } from "./authProvider";

const httpClient = async (url: string, options: fetchUtils.Options = {}) => {
  if (!options.headers) {
    options.headers = new Headers({ Accept: "application/json" });
  }

  const headers =
    options.headers instanceof Headers
      ? options.headers
      : new Headers(options.headers);

  const token = await getAccessToken();
  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
  }

  options.headers = headers;

  return fetchUtils.fetchJson(url, options);
};

const baseProvider = jsonServerProvider("/api/admin", httpClient);

const BASE_RESOURCE = "slides";

const slideResourceMap: Record<string, string> = {
  "sponsor-slides": "sponsor",
  "news-slides": "news",
  "timetable-slides": "timetable",
  "social-medias-slides": "social.media",
  "social-text-slides": "social.text",
};

interface SlideFormData {
  status?: string;
  content?: {
    type?: string;
    title?: string;
    body?: string;
  };
  author?: {
    display_name?: string;
  };
  display_options?: {
    priority?: number | string;
  };
  image_upload?: {
    rawFile?: File;
  };
  [key: string]: unknown;
}

const getActualResource = (resource: string): string => {
  return slideResourceMap[resource] ? BASE_RESOURCE : resource;
};

export const dataProvider: DataProvider = {
  ...baseProvider,

  getList: (resource, params) => {
    const slideType = slideResourceMap[resource];
    if (slideType) {
      return baseProvider.getList(BASE_RESOURCE, {
        ...params,
        filter: { ...params.filter, type: slideType },
      });
    }
    return baseProvider.getList(resource, params);
  },

  getOne: (resource, params) => {
    return baseProvider.getOne(getActualResource(resource), params);
  },

  delete: (resource, params) => {
    return baseProvider.delete(getActualResource(resource), params);
  },

  deleteMany: (resource, params) => {
    return baseProvider.deleteMany(getActualResource(resource), params);
  },

  create: (resource, params) => {
    if (getActualResource(resource) === BASE_RESOURCE) {
      return handleFileUpload(resource, params, "POST");
    }
    return baseProvider.create(resource, params);
  },

  update: (resource, params) => {
    if (getActualResource(resource) === BASE_RESOURCE) {
      return handleFileUpload(resource, params, "PUT");
    }
    return baseProvider.update(resource, params);
  },
};

const handleFileUpload = (
  resource: string,
  params: UpdateParams | CreateParams,
  method: "POST" | "PUT",
) => {
  const mappedType = slideResourceMap[resource];
  const actualResource = getActualResource(resource);

  const data = params.data as SlideFormData;
  const slideType = mappedType || data.content?.type || "slide";

  if (!data.image_upload || !data.image_upload.rawFile) {
    if (mappedType) {
      data.content = { ...data.content, type: slideType };
    }

    const newParams = { ...params, data };

    if (method === "POST") {
      return baseProvider.create(actualResource, newParams as CreateParams);
    }
    return baseProvider.update(actualResource, newParams as UpdateParams);
  }

  const formData = new FormData();

  formData.append("status", data.status || "active");
  formData.append("content.type", slideType);
  formData.append("content.title", data.content?.title || "");
  formData.append("content.body", data.content?.body || "");
  formData.append("author.display_name", data.author?.display_name || "");
  formData.append(
    "display_options.priority",
    data.display_options?.priority?.toString() || "1",
  );

  formData.append("image_upload", data.image_upload.rawFile);

  const id = "id" in params ? params.id : "";

  const url =
    method === "PUT"
      ? `/api/admin/${BASE_RESOURCE}/${id}`
      : `/api/admin/${BASE_RESOURCE}`;

  return fetch(url, {
    method: method,
    body: formData,
  })
    .then((response) => response.json())
    .then((json) => ({ data: json }));
};
