import * as pdfjs from 'pdfjs-dist';
import { TextItem, TextMarkedContent } from 'pdfjs-dist/types/src/display/api';
import { Adapter, PromiseValue } from '../types';
import { createCsvTextFromTable } from '../../utils';

/**
 * 广发银行信用卡 PDF 账单解析：
 * - 表格列固定，交易摘要可能被 PDF 拆成多个 text item；
 * - 以第一列交易日期 YYYY/MM/DD 作为新交易行开始；
 * - 第一页通常包含账单概览，需要先找到交易明细表头再开始收集。
 */
const Headers = ["交易日期", "入账日期", "交易摘要", "交易金额", "交易货币", "入账金额", "入账货币"];
const FooterStartKeywords = ["用卡安全温馨提示"];

const extractAllItemsFromPage = async (page: pdfjs.PDFPageProxy) => {
  const textContent = await page.getTextContent();
  return textContent.items.filter(
    (item: TextItem | TextMarkedContent): item is TextItem => Boolean(`${(item as TextItem)?.str ?? ''}`.trim())
  );
};

const extractHeaderInfoFromDoc = async (doc: pdfjs.PDFDocumentProxy) => {
  for (let i = 1; i <= doc.numPages; i++) {
    const page = await doc.getPage(i);
    const allItems = await extractAllItemsFromPage(page);
    const headerItems = Headers
      .map((header) => allItems.find((item) => item.str === header))
      .filter((item): item is TextItem => Boolean(item));

    if (headerItems.length === Headers.length) {
      const headerCenters = headerItems.map((item) => item.transform[4] + item.width / 2);
      const headerXRanges = headerItems.map((item, index) => ({
        title: item.str,
        colIdx: index,
        xLeft: index === 0 ? Number.NEGATIVE_INFINITY : (headerCenters[index - 1] + headerCenters[index]) / 2,
        xRight: index === headerItems.length - 1 ? Number.POSITIVE_INFINITY : (headerCenters[index] + headerCenters[index + 1]) / 2,
      }));

      return {
        headerItems,
        headerXRanges,
        headerPageNum: i,
        headerY: Math.max(...headerItems.map((item) => item.transform[5])),
      };
    }
  }

  throw Error('未找到广发信用卡交易明细表头');
};

const extractInfoFromPage = async (
  page: pdfjs.PDFPageProxy,
  pageNum: number,
  { headerXRanges, headerPageNum, headerY }: PromiseValue<ReturnType<typeof extractHeaderInfoFromDoc>>,
) => {
  const allItems = await extractAllItemsFromPage(page);

  const getItemXIndex = (item: TextItem) => {
    const xCenter = item.transform[4] + item.width / 2;
    const xRange = headerXRanges.find((r) => r.xLeft <= xCenter && xCenter < r.xRight);
    return xRange?.colIdx;
  };

  const table: string[][] = [];
  const ignoreItems: TextItem[] = [];
  let isAfterFooterStart = false;

  const isAfterHeader = (item: TextItem) => {
    return pageNum > headerPageNum || item.transform[5] < headerY;
  };

  const isTransactionDate = (item: TextItem) => {
    return getItemXIndex(item) === 0 &&
      isAfterHeader(item) &&
      /^(19|20)\d{2}\/(0[1-9]|1[0-2])\/(0[1-9]|[12]\d|3[01])$/.test(item.str);
  };

  allItems.forEach((item) => {
    if (FooterStartKeywords.some((keyword) => item.str.includes(keyword))) {
      isAfterFooterStart = true;
      ignoreItems.push(item);
      return;
    }

    if (isAfterFooterStart) {
      ignoreItems.push(item);
      return;
    }

    if (Headers.includes(item.str)) {
      return;
    }

    if (isTransactionDate(item)) {
      const newRow = Array(Headers.length).fill('');
      newRow[0] = item.str.trim();
      table.push(newRow);
      return;
    }

    const curRow = table[table.length - 1];
    if (!curRow || !isAfterHeader(item)) {
      ignoreItems.push(item);
      return;
    }

    const xIndex = getItemXIndex(item);
    if (typeof xIndex === 'undefined') {
      ignoreItems.push(item);
      return;
    }

    curRow[xIndex] = `${curRow[xIndex]}${item.str}`.trim();
  });

  return {
    ignoreItems,
    table,
  };
};

const convertFromPdf = (file: File): Promise<string> => {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();

    reader.onload = async (event) => {
      const typedArray = new Uint8Array(event.target?.result as ArrayBuffer);
      try {
        const pdf = await pdfjs.getDocument(typedArray).promise;
        const allTable: string[][] = [];

        const headerInfo = await extractHeaderInfoFromDoc(pdf);

        for (let i = headerInfo.headerPageNum; i <= pdf.numPages; i++) {
          const page = await pdf.getPage(i);
          const info = await extractInfoFromPage(page, i, headerInfo);
          allTable.push(...info.table);
        }

        const csv = createCsvTextFromTable([Headers, ...allTable]);
        resolve(csv);
      } catch (error) {
        reject(error);
      }
    };

    reader.readAsArrayBuffer(file);
  });
};

export const CgbCreditAdapter: Adapter = {
  key: 'cgb_credit',
  name: '广发银行信用卡',
  sourceFileFormat: ['pdf'],
  converter: convertFromPdf,
};
