group "default" {
  targets = ["billedapparat"]
}

target "billedapparat" {
  context    = "."
  dockerfile = "Dockerfile"
  platforms  = ["linux/amd64", "linux/arm64"]

  labels = {
    "org.opencontainers.image.url" = "https://github.com/potibm/billedapparat"
    "org.opencontainers.image.source" = "https://github.com/potibm/billedapparat"
    "org.opencontainers.image.documentation" = "https://github.com/potibm/billedapparat/tree/main/doc"
    "org.opencontainers.image.authors" = "potibm"
  }
  
  annotations = [
    "index,manifest:org.opencontainers.image.title=Billedapparat",
    "index,manifest:org.opencontainers.image.description=A party information system for the bigscreen at demoparties.",
    "index,manifest:org.opencontainers.image.url=https://github.com/potibm/billedapparat",
    "index,manifest:org.opencontainers.image.source=https://github.com/potibm/billedapparat",
    "index,manifest:org.opencontainers.image.documentation=https://github.com/potibm/billedapparat/tree/main/doc",
    "index,manifest:org.opencontainers.image.licenses=MIT",
    "index,manifest:org.opencontainers.image.authors=potibm"
  ]
}
