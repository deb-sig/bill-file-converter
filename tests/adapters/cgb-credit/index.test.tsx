import { vi, beforeEach, test, describe, expect } from 'vitest';

import { CgbCreditAdapter } from '../../../src/adapters/cgb-credit';

import p1TextItems from './p1.json';
import resultCsvText from './result.csv?raw';

const pages = [p1TextItems];

describe('adapter for cgb credit', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  test('converter', async () => {
    vi.mock('pdfjs-dist', () => ({
      getDocument: vi.fn(() => ({
        promise: Promise.resolve({
          numPages: pages.length,
          getPage: (pageNum: number) => ({
            getTextContent: () => Promise.resolve({
              items: pages[pageNum - 1] || null,
            })
          })
        })
      }))
    }));

    const mockFile = new File([], 'cgb_credit.pdf');
    const csvText = await CgbCreditAdapter.converter(mockFile);

    expect(csvText).toEqual(resultCsvText.trimEnd());
  });
});
