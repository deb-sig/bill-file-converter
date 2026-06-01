import PostalMime from 'postal-mime';
import { Adapter } from '../types';
import { createCsvTextFromTable } from '../../utils';

const Headers = ['交易日期', '记账日期', '交易说明', '交易币种/金额', '入账币种/金额'];

// 标准化单元格文本，兼容邮件模板中的换行、缩进和不换行空格。
const getNormalizedText = ($element: Element) => {
  const $clone = $element.cloneNode(true) as Element;
  $clone.querySelectorAll('br').forEach(($br) => $br.replaceWith(' '));
  return ($clone.textContent || '').replace(/\u00a0/g, ' ').trim().replace(/\s+/g, ' ');
};

const getNormalizedHeaderText = ($element: Element) => {
  return getNormalizedText($element).replace(/\s+/g, '');
};

const getNormalizedCellText = ($element: Element) => {
  return getNormalizedText($element)
    .replace(/(\d{4})\s*-\s*(\d{2})\s*-\s*(\d{2})/g, '$1-$2-$3');
};

const extractInfoFromHtml = (html: string) => {
  const parser = new DOMParser();
  const $doc = parser.parseFromString(html, 'text/html');
  const $tableList = $doc.querySelectorAll('table#table3');

  const resultTable: string[][] = [Headers];

  $tableList.forEach(($table) => {
    const $trList = [...$table.querySelectorAll('tr')];
    const headerRowIndex = $trList.findIndex(($tr) => {
      const $cellList = [...$tr.querySelectorAll('th,td')];
      return $cellList.length === Headers.length &&
        $cellList.every(($cell, idx) => getNormalizedHeaderText($cell) === Headers[idx]);
    });
    if (headerRowIndex < 0) {
      return;
    }

    $trList.slice(headerRowIndex + 1).forEach(($tr) => {
      const $tdList = [...$tr.querySelectorAll('td')];
      
      if ($tdList.length !== Headers.length) {
        return;
      }

      const row = $tdList.map(getNormalizedCellText);
      resultTable.push(row);
    })
  });

  return resultTable;
}

const convertFromEml = (file: File): Promise<string> => {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();

    reader.onload = async (event) => {
      try {
        const typedArray = new Uint8Array(event.target?.result as ArrayBuffer);
        const email = await PostalMime.parse(typedArray);
        const table = extractInfoFromHtml(email.html || "");
        const csv = createCsvTextFromTable(table);
        resolve(csv);
      } catch (error) {
        reject(error);
      }
    };

    reader.readAsArrayBuffer(file);
  });
};

export const BocomCreditAdapter: Adapter = {
  key: 'bocom_credit',
  name: '交通银行信用卡',
  sourceFileFormat: ['eml'],
  converter: convertFromEml,
}
