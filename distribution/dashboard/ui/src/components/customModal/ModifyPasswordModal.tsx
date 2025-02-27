import { getNSName } from '@/pages/Cluster/Detail/Overview/helper';
import { changeTenantPassword } from '@/services/tenant';
import { intl } from '@/utils/intl';
import { Form, Input, message } from 'antd';
import type { CommonModalType } from '.';
import CustomModal from '.';

type FieldType = {
  Password: string;
};

export default function ModifyPasswordModal({
  visible,
  setVisible,
  successCallback,
}: CommonModalType) {
  const [form] = Form.useForm();

  const handleSubmit = async () => {
    try {
      await form.validateFields();
      form.submit();
    } catch (err) {}
  };

  const handleCancel = () => setVisible(false);
  const onFinish = async (values: any) => {
    const [namespace, name] = getNSName();
    const res = await changeTenantPassword({
      ns: namespace,
      name,
      User: 'root',
      Password: values.password,
    });
    if (res.successful) {
      message.success(res.message);
      successCallback();
      form.resetFields();
      setVisible(false);
    }
  };
  return (
    <CustomModal
      title={intl.formatMessage({
        id: 'Dashboard.components.customModal.ModifyPasswordModal.ModifyRootPassword',
        defaultMessage: '修改 root 密码',
      })}
      isOpen={visible}
      handleOk={handleSubmit}
      handleCancel={handleCancel}
    >
      <Form
        form={form}
        onFinish={onFinish}
        style={{ maxWidth: 600 }}
        autoComplete="off"
      >
        <Form.Item<FieldType>
          label={intl.formatMessage({
            id: 'Dashboard.components.customModal.ModifyPasswordModal.EnterANewPassword',
            defaultMessage: '输入新密码',
          })}
          name="password"
          rules={[
            {
              required: true,
              message: intl.formatMessage({
                id: 'Dashboard.components.customModal.ModifyPasswordModal.PleaseEnter',
                defaultMessage: '请输入',
              }),
            },
          ]}
        >
          <Input.Password
            placeholder={intl.formatMessage({
              id: 'Dashboard.components.customModal.ModifyPasswordModal.PleaseEnter',
              defaultMessage: '请输入',
            })}
          />
        </Form.Item>
        <Form.Item<FieldType>
          label={intl.formatMessage({
            id: 'Dashboard.components.customModal.ModifyPasswordModal.EnterAgain',
            defaultMessage: '再次输入',
          })}
          name="passwordAgain"
          validateTrigger="onBlur"
          rules={[
            {
              required: true,
              message: intl.formatMessage({
                id: 'Dashboard.components.customModal.ModifyPasswordModal.PleaseEnter',
                defaultMessage: '请输入',
              }),
            },
            () => ({
              validator(_: any, value: string) {
                if (
                  form.getFieldValue('password') &&
                  value !== form.getFieldValue('password')
                ) {
                  return Promise.reject(
                    new Error(
                      intl.formatMessage({
                        id: 'Dashboard.components.customModal.ModifyPasswordModal.TheTwoInputsAreInconsistent',
                        defaultMessage: '两次输入不一致',
                      }),
                    ),
                  );
                }
                return Promise.resolve();
              },
            }),
          ]}
        >
          <Input.Password
            placeholder={intl.formatMessage({
              id: 'Dashboard.components.customModal.ModifyPasswordModal.PleaseEnter',
              defaultMessage: '请输入',
            })}
          />
        </Form.Item>
      </Form>
    </CustomModal>
  );
}
