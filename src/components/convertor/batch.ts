import { Adapter } from '../../adapters/types';

export interface ConvertSuccess {
  sourceFile: File;
  csv: string;
}

export interface ConvertFailure {
  sourceFile: File;
  message: string;
}

export interface ConvertFilesResult {
  successes: ConvertSuccess[];
  failures: ConvertFailure[];
}

const getErrorMessage = (error: unknown) => {
  return error instanceof Error ? error.message : String(error);
};

/**
 * 逐个转换文件，单个文件失败不影响后续文件。
 */
export const convertFiles = async (adapter: Adapter, files: File[]): Promise<ConvertFilesResult> => {
  const successes: ConvertSuccess[] = [];
  const failures: ConvertFailure[] = [];

  for (const file of files) {
    try {
      const csv = await adapter.converter(file);
      successes.push({ sourceFile: file, csv });
    } catch (error) {
      failures.push({ sourceFile: file, message: getErrorMessage(error) });
    }
  }

  return { successes, failures };
};
