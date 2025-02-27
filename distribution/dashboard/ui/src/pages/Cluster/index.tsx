import { PageContainer } from '@ant-design/pro-components';
import { useNavigate } from '@umijs/max';
import { useRequest } from 'ahooks';
import { Row } from 'antd';
import { useState } from 'react';

import EventsTable from '@/components/EventsTable';
import MonitorComp from '@/components/MonitorComp';
import ClusterList from './ClusterList';
// import Monitor from './Monitor';
import { getObclusterListReq } from '@/services';
import type { LabelType, QueryRangeType } from './Detail/Monitor';

const defaultQueryRange:QueryRangeType = {
  step: 20,
  endTimestamp: Math.floor(new Date().valueOf() / 1000),
  startTimestamp: Math.floor(new Date().valueOf() / 1000) - 60 * 30,
}

//集群概览页
const ClusterPage: React.FC = () => {
  const navigate = useNavigate();
  const [clusterNames, setClusterNames] = useState<LabelType[]>([]);

  const { data: clusterList } = useRequest(getObclusterListReq, {
    onSuccess: (data) => {
      let clusterNames: LabelType[] = data.map((item) => ({
        key: 'ob_cluster_name',
        value: item.clusterName,
      }));
      setClusterNames(clusterNames);
    },
  });

  const handleAddCluster = () => navigate('new');

  return (
    <PageContainer>
      <Row gutter={[16, 16]}>
        <ClusterList
          clusterList={clusterList}
          handleAddCluster={handleAddCluster}
        />
        <EventsTable objectType="OBCLUSTER" />
      </Row>
      <MonitorComp 
        filterLabel={clusterNames} 
        queryScope='OBCLUSTER_OVERVIEW' 
        type='OVERVIEW' 
        queryRange={defaultQueryRange}/>
    </PageContainer>
  );
};

export default ClusterPage;
