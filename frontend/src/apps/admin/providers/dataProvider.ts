// src/apps/admin/dataProvider.ts
import jsonServerProvider from "ra-data-json-server";

const baseProvider = jsonServerProvider("/api/admin");

export const dataProvider = {
  ...baseProvider,

  // 1. LISTE: /sponsors -> /slides?type=sponsor
  getList: (resource, params) => {
    if (resource === "sponsors") {
      // Wir leiten die Anfrage an 'slides' um und schmuggeln den Filter rein
      return baseProvider.getList("slides", {
        ...params,
        filter: { ...params.filter, type: "sponsor" },
      });
    }
    return baseProvider.getList(resource, params);
  },

  // 2. EINZEL-EINTRAG (für den Edit-View)
  getOne: (resource, params) => {
    if (resource === "sponsors") {
      return baseProvider.getOne("slides", params);
    }
    return baseProvider.getOne(resource, params);
  },

  // 3. LÖSCHEN
  delete: (resource, params) => {
    if (resource === "sponsors") {
      return baseProvider.delete("slides", params);
    }
    return baseProvider.delete(resource, params);
  },

  deleteMany: (resource, params) => {
    if (resource === "sponsors") {
      return baseProvider.deleteMany("slides", params);
    }
    return baseProvider.deleteMany(resource, params);
  },

  // Wir biegen sowohl create als auch update um
  create: (resource: string, params: any) => {
    if (resource !== "slides" && resource !== "sponsors")
      return baseProvider.create(resource, params);
    return handleFileUpload(resource, params, "POST");
  },
  update: (resource: string, params: any) => {
    if (resource !== "slides" && resource !== "sponsors")
      return baseProvider.update(resource, params);
    return handleFileUpload(resource, params, "PUT");
  },
};

// Hilfsfunktion für den Multipart-Upload
const handleFileUpload = (
  resource: string,
  params: any,
  method: "POST" | "PUT",
) => {
  // Falls kein Bild da ist, machen wir normales JSON
  if (!params.data.image_upload || !params.data.image_upload.rawFile) {
    if (method === "POST") return baseProvider.create(resource, params);
    return baseProvider.update(resource, params);
  }

  const formData = new FormData();

  // Wir klopfen die verschachtelten Objekte flach, damit Go sie einfach lesen kann
  formData.append("status", params.data.status || "active");
  formData.append(
    "content.type",
    params.data.content?.type || (resource === "sponsors" ? "sponsor" : "news"),
  );
  formData.append("content.text", params.data.content?.text || "");
  formData.append(
    "author.display_name",
    params.data.author?.display_name || "",
  );
  formData.append(
    "display_options.priority",
    params.data.display_options?.priority?.toString() || "1",
  );

  // Die Datei selbst
  formData.append("image_upload", params.data.image_upload.rawFile);

  const url =
    method === "PUT"
      ? `/api/admin/slides/${params.id}` // Wir schicken alles an /slides
      : `/api/admin/slides`;

  return fetch(url, {
    method: method,
    body: formData,
  })
    .then((response) => response.json())
    .then((json) => ({ data: json }));
};
