export const compareKey = (a: string, b: string) => {
  const lowerA = a.toLowerCase();
  const lowerB = b.toLowerCase();
  if (lowerA < lowerB) {
      return -1;
  }
  if (lowerA > lowerB) {
      return 1;
  }
  return 0;
};

export const createCsvTextFromTable = (table: string[][]) => {
  const csv = table
    // 如果单元格内容包含逗号，则用双引号包裹
    .map((row) => row.map((cell) => cell.includes(',') ? `"${cell.replace(/"/g, '""')}"` : cell).join(','))
    .join('\n');
  return csv;
}

export interface ZipFileEntry {
  name: string;
  content: string;
}

const textEncoder = new TextEncoder();

const makeCrc32Table = () => {
  const table = new Uint32Array(256);
  for (let i = 0; i < table.length; i++) {
    let crc = i;
    for (let j = 0; j < 8; j++) {
      crc = crc & 1 ? 0xedb88320 ^ (crc >>> 1) : crc >>> 1;
    }
    table[i] = crc >>> 0;
  }
  return table;
};

const crc32Table = makeCrc32Table();

const getCrc32 = (bytes: Uint8Array) => {
  let crc = 0xffffffff;
  for (const byte of bytes) {
    crc = crc32Table[(crc ^ byte) & 0xff] ^ (crc >>> 8);
  }
  return (crc ^ 0xffffffff) >>> 0;
};

const getDosDateTime = (date = new Date()) => {
  const year = Math.max(date.getFullYear(), 1980);
  const dosTime = (date.getHours() << 11) | (date.getMinutes() << 5) | Math.floor(date.getSeconds() / 2);
  const dosDate = ((year - 1980) << 9) | ((date.getMonth() + 1) << 5) | date.getDate();
  return { dosDate, dosTime };
};

const getZipFileName = (name: string) => {
  return name.endsWith('.csv') ? name : `${name}.csv`;
};

const appendUint16 = (target: number[], value: number) => {
  target.push(value & 0xff, (value >>> 8) & 0xff);
};

const appendUint32 = (target: number[], value: number) => {
  target.push(value & 0xff, (value >>> 8) & 0xff, (value >>> 16) & 0xff, (value >>> 24) & 0xff);
};

const appendBytes = (target: number[], bytes: Uint8Array) => {
  target.push(...bytes);
};

/**
 * 使用 ZIP store 模式打包多个文本文件，避免浏览器拦截连续多文件下载。
 */
export const createZipBlobFromFiles = (files: ZipFileEntry[]) => {
  const zipBytes: number[] = [];
  const centralDirectoryBytes: number[] = [];
  const { dosDate, dosTime } = getDosDateTime();

  files.forEach((file) => {
    const fileNameBytes = textEncoder.encode(getZipFileName(file.name));
    const contentBytes = textEncoder.encode(file.content);
    const crc32 = getCrc32(contentBytes);
    const localHeaderOffset = zipBytes.length;

    appendUint32(zipBytes, 0x04034b50);
    appendUint16(zipBytes, 20);
    appendUint16(zipBytes, 0x0800);
    appendUint16(zipBytes, 0);
    appendUint16(zipBytes, dosTime);
    appendUint16(zipBytes, dosDate);
    appendUint32(zipBytes, crc32);
    appendUint32(zipBytes, contentBytes.length);
    appendUint32(zipBytes, contentBytes.length);
    appendUint16(zipBytes, fileNameBytes.length);
    appendUint16(zipBytes, 0);
    appendBytes(zipBytes, fileNameBytes);
    appendBytes(zipBytes, contentBytes);

    appendUint32(centralDirectoryBytes, 0x02014b50);
    appendUint16(centralDirectoryBytes, 20);
    appendUint16(centralDirectoryBytes, 20);
    appendUint16(centralDirectoryBytes, 0x0800);
    appendUint16(centralDirectoryBytes, 0);
    appendUint16(centralDirectoryBytes, dosTime);
    appendUint16(centralDirectoryBytes, dosDate);
    appendUint32(centralDirectoryBytes, crc32);
    appendUint32(centralDirectoryBytes, contentBytes.length);
    appendUint32(centralDirectoryBytes, contentBytes.length);
    appendUint16(centralDirectoryBytes, fileNameBytes.length);
    appendUint16(centralDirectoryBytes, 0);
    appendUint16(centralDirectoryBytes, 0);
    appendUint16(centralDirectoryBytes, 0);
    appendUint16(centralDirectoryBytes, 0);
    appendUint32(centralDirectoryBytes, 0);
    appendUint32(centralDirectoryBytes, localHeaderOffset);
    appendBytes(centralDirectoryBytes, fileNameBytes);
  });

  const centralDirectoryOffset = zipBytes.length;
  appendBytes(zipBytes, new Uint8Array(centralDirectoryBytes));

  appendUint32(zipBytes, 0x06054b50);
  appendUint16(zipBytes, 0);
  appendUint16(zipBytes, 0);
  appendUint16(zipBytes, files.length);
  appendUint16(zipBytes, files.length);
  appendUint32(zipBytes, centralDirectoryBytes.length);
  appendUint32(zipBytes, centralDirectoryOffset);
  appendUint16(zipBytes, 0);

  return new Blob([new Uint8Array(zipBytes)], { type: 'application/zip' });
};

export const downloadCsvFile = (csv: string, name: string = 'output') => {
  const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = `${name}.csv`;
  a.click();
  URL.revokeObjectURL(url);
}

export const downloadZipFile = (files: ZipFileEntry[], name: string = 'output') => {
  const blob = createZipBlobFromFiles(files);
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = `${name}.zip`;
  a.click();
  URL.revokeObjectURL(url);
}
