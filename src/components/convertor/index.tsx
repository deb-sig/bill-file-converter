import { App as AntdApp, Button, Segmented, Upload } from "antd";
import React, { useRef, useState } from "react";
import { SegmentedOptions } from "antd/es/segmented";
import { RcFile } from "antd/es/upload";
import { CheckCircleOutlined, ClockCircleOutlined, CloudUploadOutlined, DownloadOutlined, LoadingOutlined } from "@ant-design/icons";
import { AdapterMap, AdapterList } from '../../adapters';
import { downloadCsvFile, downloadZipFile } from "../../utils";
import { convertFiles, ConvertFailure, ConvertSuccess } from "./batch";
import './index.css';

const options: SegmentedOptions<string> = AdapterList.map((a) => {
  return (
    {
      label: (
        <div style={{ padding: 4 }}>
          <div>{a.name}</div>
        </div>
      ),
      value: a.key,
    }
  )
});

const Convertor: React.FC = () => {
  const { modal } = AntdApp.useApp();
  const [selectedKey, setSelectedKey] = useState<string>(AdapterList[0]?.key);
  const [sourceFiles, setSourceFiles] = useState<File[]>([]);
  const [convertedFiles, setConvertedFiles] = useState<ConvertSuccess[]>([]);
  const [failedFiles, setFailedFiles] = useState<ConvertFailure[]>([]);
  const [converting, setConverting] = useState(false);
  const uploadFileListRef = useRef<File[]>([]);
  const uploadTimerRef = useRef<number | undefined>(undefined);
  const selectedAdapter = AdapterMap[selectedKey];

  const resetState = () => {
    setSourceFiles([]);
    setConvertedFiles([]);
    setFailedFiles([]);
    setConverting(false);
    uploadFileListRef.current = [];
  };

  const handleSelectorChange = (k: string) => {
    setSelectedKey(k);
    resetState();
  }

  const handleBatchUpload = async (files: File[]) => {
    setSourceFiles(files);
    setConvertedFiles([]);
    setFailedFiles([]);
    setConverting(true);

    const result = await convertFiles(selectedAdapter, files);
    setConvertedFiles(result.successes);
    setFailedFiles(result.failures);
    setConverting(false);

    if (result.failures.length) {
      modal.warning({
        title: '部分文件解析失败',
        content: (
          <div>
            {result.failures.map((failure) => (
              <div key={failure.sourceFile.name}>
                {failure.sourceFile.name}：{failure.message}
              </div>
            ))}
          </div>
        ),
      });
    }
  };

  const handleUpload = (_file: RcFile, fileList: RcFile[]) => {
    uploadFileListRef.current = fileList;
    if (uploadTimerRef.current) {
      window.clearTimeout(uploadTimerRef.current);
    }
    uploadTimerRef.current = window.setTimeout(() => {
      void handleBatchUpload([...uploadFileListRef.current]);
      uploadTimerRef.current = undefined;
    }, 0);

    return false;
  };

  const handleDownload = () => {
    if (convertedFiles.length === 1) {
      const [{ sourceFile, csv }] = convertedFiles;
      downloadCsvFile(csv, sourceFile.name);
      return;
    }

    downloadZipFile(
      convertedFiles.map(({ sourceFile, csv }) => ({ name: sourceFile.name, content: csv })),
      `${selectedAdapter.key}-csv`,
    );
  }

  const renderStatus = () => {
    if (converting) {
      return <><LoadingOutlined /> 文件转换中：共 {sourceFiles.length} 个文件</>;
    }

    if (sourceFiles.length) {
      return (
        <>
          <CheckCircleOutlined />
          文件转换完成：成功 {convertedFiles.length} 个，失败 {failedFiles.length} 个
        </>
      );
    }

    return <><ClockCircleOutlined /> 等待文件上传</>;
  }

  return (
    <div className="app-convertor">
      <Segmented<string>
        className="app-convertor-selector"
        options={options}
        value={selectedKey}
        onChange={handleSelectorChange}
      />
      <div className="app-convertor-files">
        <Upload.Dragger
          className="app-convertor-upload"
          name="file"
          multiple
          showUploadList={false}
          accept={selectedAdapter.sourceFileFormat.map((f) => `.${f}`).join(',')}
          beforeUpload={handleUpload}
        >
          <CloudUploadOutlined className="app-convertor-upload-icon" />
          <p className="ant-upload-text">点击或拖拽文件到此处上传</p>
          <p className="ant-upload-hint">支持的格式：{selectedAdapter.sourceFileFormat.join('/')}</p>
        </Upload.Dragger>
        <div className="app-convertor-download">
          <div className="app-convertor-download-status">
            {renderStatus()}
          </div>
          <Button
            type="primary"
            icon={<DownloadOutlined />}
            disabled={converting || !convertedFiles.length}
            onClick={handleDownload}
          >
            {convertedFiles.length > 1 ? `下载 ZIP (${convertedFiles.length})` : '下载 CSV'}
          </Button>
        </div>
      </div>
    </div>
  )
}

export default Convertor;
