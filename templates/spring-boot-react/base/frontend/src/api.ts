export type Item = {
  id: number;
  name: string;
  createdAt: string;
};

const baseUrl = "/api";

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${baseUrl}${path}`, {
    headers: {
      "Content-Type": "application/json",
      ...init?.headers,
    },
    ...init,
  });

  if (!response.ok) {
    throw new Error(`Request failed with status ${response.status}`);
  }

  if (response.status === 204) {
    return undefined as T;
  }

  return (await response.json()) as T;
}

export function listItems(): Promise<Item[]> {
  return request<Item[]>("/items");
}

export function createItem(name: string): Promise<Item> {
  return request<Item>("/items", {
    method: "POST",
    body: JSON.stringify({ name }),
  });
}

export function deleteItem(id: number): Promise<void> {
  return request<void>(`/items/${id}`, { method: "DELETE" });
}
