import { intl } from '@/utils/intl';
import { useRequest } from 'ahooks';
import { Form, Input, message } from 'antd';

import { createNameSpace } from '@/services';
import type { CommonModalType } from '.';
import CustomModal from '.';

export default function AddNSModal({
  visible,
  setVisible,
  successCallback,
}: CommonModalType) {
  const { run: createNS } = useRequest(createNameSpace, {
    manual: true,
    onSuccess: ({ successful }) => {
      if (successful) {
        message.success(
          intl.formatMessage({
            id: 'OBDashboard.components.customModal.AddNSModal.AddedSuccessfully',
            defaultMessage: '新增成功',
          }),
        );
        setVisible(false);
        successCallback();
      }
    },
  });
  const [form] = Form.useForm();
  const handleSubmit = async () => {
    try {
      await form.validateFields();
      form.submit();
    } catch (err) {}
  };
  const handleCancel = () => setVisible(false);
  const onFinish = async (val: any) => {
    await createNS(val.namespace);
  };
  return (
    <CustomModal
      title={intl.formatMessage({
        id: 'OBDashboard.components.customModal.AddNSModal.AddNamespace',
        defaultMessage: '新增命名空间',
      })}
      isOpen={visible}
      handleOk={handleSubmit}
      handleCancel={handleCancel}
    >
      <Form form={form} onFinish={onFinish}>
        <Form.Item
          label="namespace"
          name="namespace"
          rules={[
            {
              required: true,
              message: intl.formatMessage({
                id: 'OBDashboard.components.customModal.AddNSModal.EnterNamespace',
                defaultMessage: '请输入namespace!',
              }),
            },
          ]}
        >
          <Input
            placeholder={intl.formatMessage({
              id: 'OBDashboard.components.customModal.AddNSModal.PleaseEnter',
              defaultMessage: '请输入',
            })}
          />
        </Form.Item>
      </Form>
    </CustomModal>
  );
}
