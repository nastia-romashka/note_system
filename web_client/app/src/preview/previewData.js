export const PREVIEW_DATA = {
  categories: [
    {
      uuid: "cat-1",
      name: "Категория 1",
      color: "#9db8ff",
      children: [
        {
          uuid: "cat-1-1",
          name: "1.1 Подкатегория",
          color: "#8ed7d1",
          children: [
            {
              uuid: "cat-1-1-1",
              name: "1.1.1 Подкатегория",
              color: "#c7b2ff",
            },
          ],
        },
      ],
    },
    { uuid: "cat-2", name: "Категория 2", color: "#9fd7b2" },
  ],
  tags: [
    { uuid: "tag-1", name: "tag1" },
    { uuid: "tag-2", name: "tag2" },
    { uuid: "tag-3", name: "tag3" },
  ],
  notes: {
    "cat-1": [
      {
        uuid: "note-1",
        header: "Имя заметки",
        body:
          "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.",
        short_body:
          "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.",
        category_uuid: "cat-1",
        tags: ["tag-1", "tag-2", "tag-3"],
        created_date: 1780226400,
        updated_at: 1780312800,
        event: {
          enabled: true,
          start_at: 1780402800,
          end_at: 1780406400,
        },
      },
    ],
    "cat-1-1": [
      {
        uuid: "note-2",
        header: "Заметка по подкатегории",
        body:
          "Структурируй мысли по вложенным категориям и собирай связанные материалы в одном месте. Такой режим помогает посмотреть будущий интерфейс без запуска backend.",
        short_body:
          "Структурируй мысли по вложенным категориям и собирай связанные материалы в одном месте.",
        category_uuid: "cat-1-1",
        tags: ["tag-2"],
        created_date: 1780140000,
        updated_at: 1780312800,
        event: {
          enabled: true,
          start_at: 1780489200,
          end_at: 1780492800,
        },
      },
    ],
    "cat-1-1-1": [],
    "cat-2": [
      {
        uuid: "note-3",
        header: "Вторая категория",
        body:
          "Отдельная область для рабочих заметок, быстрых черновиков и небольших материалов, которые нужны в течение дня.",
        short_body:
          "Отдельная область для рабочих заметок, быстрых черновиков и небольших материалов.",
        category_uuid: "cat-2",
        tags: [],
        created_date: 1780053600,
        updated_at: 1780312800,
      },
    ],
  },
  files: {
    "note-1": [
      { id: "file-1", name: "image.png", size: 183421, content_type: "image/png" },
      { id: "file-2", name: "brief.pdf", size: 94213, content_type: "application/pdf" },
      { id: "file-3", name: "Link.txt", size: 1024, content_type: "text/plain" },
    ],
    "note-2": [
      {
        id: "file-4",
        name: "roadmap.docx",
        size: 42018,
        content_type: "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
      },
    ],
    "note-3": [],
  },
};
