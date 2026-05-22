import { describe, expect, test } from 'vitest';

import { createZipBlobFromFiles } from '../../src/utils';

const readBlobAsArrayBuffer = (blob: Blob) => {
  return new Promise<ArrayBuffer>((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => resolve(reader.result as ArrayBuffer);
    reader.onerror = () => reject(reader.error);
    reader.readAsArrayBuffer(blob);
  });
};

describe('createZipBlobFromFiles', () => {
  test('把多个 CSV 内容打包成一个包含独立文件的 zip', async () => {
    const blob = createZipBlobFromFiles([
      { name: 'first.csv', content: 'a,b\n1,2' },
      { name: 'second.csv', content: 'c,d\n3,4' },
    ]);

    const bytes = new Uint8Array(await readBlobAsArrayBuffer(blob));
    const text = new TextDecoder().decode(bytes);
    const view = new DataView(bytes.buffer);
    const eocdOffset = bytes.length - 22;

    expect(blob.type).toBe('application/zip');
    expect(text).toContain('first.csv');
    expect(text).toContain('second.csv');
    expect(view.getUint32(0, true)).toBe(0x04034b50);
    expect(view.getUint32(eocdOffset, true)).toBe(0x06054b50);
    expect(view.getUint16(eocdOffset + 10, true)).toBe(2);
  });
});
