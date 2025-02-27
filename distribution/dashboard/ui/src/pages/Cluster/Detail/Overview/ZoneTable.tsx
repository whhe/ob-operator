import { intl } from '@/utils/intl'; //@ts-nocheck
import { ProCard } from '@ant-design/pro-components';
import { Col, Table, Tag, message } from 'antd';
import type { ColumnType } from 'antd/es/table';

import showDeleteConfirm from '@/components/customModal/DeleteModal';
import { COLOR_MAP } from '@/constants';
import { deleteObzone } from '@/services';
import { getNSName } from './helper';

interface ZoneTableProps {
  zones: API.Zone[];
  setVisible: React.Dispatch<React.SetStateAction<boolean>>;
  chooseZoneRef: React.MutableRefObject<string>;
  typeRef: React.MutableRefObject<API.ModalType>;
  setChooseServerNum: React.Dispatch<React.SetStateAction<number>>;
}

export default function ZoneTable({
  zones,
  setVisible,
  chooseZoneRef,
  typeRef,
  setChooseServerNum,
}: ZoneTableProps) {
  const getZoneColumns = (remove, clickScale) => {
    const columns: ColumnType<API.Zone> = [
      {
        title: intl.formatMessage({
          id: 'OBDashboard.Detail.Overview.ZoneTable.ZoneName',
          defaultMessage: 'Zone名',
        }),
        dataIndex: 'name',
        key: 'name',
        width: 190,
      },
      {
        title: intl.formatMessage({
          id: 'OBDashboard.Detail.Overview.ZoneTable.Namespace',
          defaultMessage: '命名空间',
        }),
        dataIndex: 'namespace',
        key: 'namespace',
      },
      {
        title: 'zone',
        dataIndex: 'zone',
        key: 'zone',
      },
      {
        title: intl.formatMessage({
          id: 'OBDashboard.Detail.Overview.ZoneTable.NumberOfMachines',
          defaultMessage: '机器数量',
        }),
        dataIndex: 'replicas',
        key: 'replicas',
      },
      {
        title: intl.formatMessage({
          id: 'OBDashboard.Detail.Overview.ZoneTable.RootServiceIp',
          defaultMessage: '根服务IP',
        }),
        dataIndex: 'rootService',
        key: 'rootService',
      },
      {
        title: intl.formatMessage({
          id: 'OBDashboard.Detail.Overview.ZoneTable.Status',
          defaultMessage: '状态',
        }),
        dataIndex: 'status',
        key: 'status',
        render: (value) => <Tag color={COLOR_MAP.get(value)}>{value} </Tag>,
      },
      {
        title: intl.formatMessage({
          id: 'OBDashboard.Detail.Overview.ZoneTable.Operation',
          defaultMessage: '操作',
        }),
        key: 'action',
        render: (value, record) => {
          return (
            <>
              <a
                style={{ marginRight: 10 }}
                onClick={() => {
                  clickScale(record.zone);
                  setChooseServerNum(record.replicas);
                }}
              >
                {intl.formatMessage({
                  id: 'OBDashboard.Detail.Overview.ZoneTable.Expansion',
                  defaultMessage: '扩缩容',
                })}
              </a>
              <a
                onClick={() => {
                  showDeleteConfirm({
                    onOk: () => remove(record.zone),
                    title: intl.formatMessage({
                      id: 'OBDashboard.Detail.Overview.ZoneTable.AreYouSureYouWant',
                      defaultMessage: '你确定要删除该zone吗？',
                    }),
                  });
                }}
              >
                {intl.formatMessage({
                  id: 'OBDashboard.Detail.Overview.ZoneTable.Delete',
                  defaultMessage: '删除',
                })}
              </a>
            </>
          );
        },
      },
    ];

    return columns;
  };
  const clickScale = (zoneName: string) => {
    chooseZoneRef.current = zoneName;
    typeRef.current = 'scaleServer';
    setVisible(true);
  };
  //删除的ns和name是集群的
  const handleDelete = async (zoneName:string) => {
    const [ns, name] = getNSName();
    const res = await deleteObzone({
      ns,
      name,
      zoneName,
    });
    if (res.successful) {
      message.success(
        intl.formatMessage({
          id: 'OBDashboard.Detail.Overview.ZoneTable.OperationSucceeded',
          defaultMessage: '操作成功！',
        }),
      );
    }
  };
  return (
    <Col span={24}>
      <ProCard>
        <Table
          rowKey="name"
          pagination={{simple:true}}
          columns={getZoneColumns(handleDelete, clickScale)}
          dataSource={zones}
        />
      </ProCard>
    </Col>
  );
}
