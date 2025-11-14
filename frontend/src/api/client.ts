const API_BASE_URL = import.meta.env.VITE_API_URL || '/api';

export interface CreatePasteRequest {
  content: string;
  tone: string;
}

export interface CreatePasteResponse {
  paste_id: string;
  original_language: string;
  available_languages: string[];
}

export interface GetPasteResponse {
  paste_id: string;
  original_language: string;
  tone: string;
  created_at: string;
  original: string;
  translations: { [key: string]: string };
  available_translations: string[];
}

export interface TranslateResponse {
  language: string;
  translation: string;
}

class APIClient {
  async createPaste(request: CreatePasteRequest): Promise<CreatePasteResponse> {
    const response = await fetch(`${API_BASE_URL}/pastes`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(request),
    });

    if (!response.ok) {
      const error = await response.text();
      throw new Error(error || 'Failed to create paste');
    }

    return response.json();
  }

  async getPaste(pasteId: string): Promise<GetPasteResponse> {
    const response = await fetch(`${API_BASE_URL}/pastes/${pasteId}`);

    if (!response.ok) {
      const error = await response.text();
      throw new Error(error || 'Failed to get paste');
    }

    return response.json();
  }

  async translate(pasteId: string, language: string): Promise<TranslateResponse> {
    const response = await fetch(`${API_BASE_URL}/pastes/${pasteId}/translate?lang=${language}`);

    if (!response.ok) {
      const error = await response.text();
      throw new Error(error || 'Failed to translate');
    }

    return response.json();
  }
}

export const apiClient = new APIClient();
