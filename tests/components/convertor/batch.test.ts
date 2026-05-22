import { describe, expect, test } from 'vitest';

import { convertFiles } from '../../../src/components/convertor/batch';
import { Adapter } from '../../../src/adapters/types';

describe('convertFiles', () => {
  test('逐个转换文件并保留每个文件对应的独立 CSV 结果', async () => {
    const adapter: Adapter = {
      key: 'mock',
      name: '模拟账单',
      sourceFileFormat: ['pdf'],
      converter: async (file) => `csv:${file.name}`,
    };
    const files = [
      new File([], 'first.pdf'),
      new File([], 'second.pdf'),
    ];

    const result = await convertFiles(adapter, files);

    expect(result.successes).toEqual([
      { sourceFile: files[0], csv: 'csv:first.pdf' },
      { sourceFile: files[1], csv: 'csv:second.pdf' },
    ]);
    expect(result.failures).toEqual([]);
  });

  test('单个文件转换失败时继续转换其他文件并记录失败原因', async () => {
    const adapter: Adapter = {
      key: 'mock',
      name: '模拟账单',
      sourceFileFormat: ['pdf'],
      converter: async (file) => {
        if (file.name === 'bad.pdf') {
          throw new Error('解析失败');
        }
        return `csv:${file.name}`;
      },
    };
    const files = [
      new File([], 'good.pdf'),
      new File([], 'bad.pdf'),
      new File([], 'next.pdf'),
    ];

    const result = await convertFiles(adapter, files);

    expect(result.successes).toEqual([
      { sourceFile: files[0], csv: 'csv:good.pdf' },
      { sourceFile: files[2], csv: 'csv:next.pdf' },
    ]);
    expect(result.failures).toEqual([
      { sourceFile: files[1], message: '解析失败' },
    ]);
  });
});
