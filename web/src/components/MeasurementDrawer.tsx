// web/src/components/MeasurementDrawer.tsx
import { useEffect } from 'react'
import {
  Drawer,
  Form,
  InputNumber,
  DatePicker,
  Input,
  Button,
  message,
} from 'antd'
import {
  createMeasurement,
  updateMeasurement,
} from '../api/measurements'
import type { Measurement } from '../api/measurements'
import dayjs from 'dayjs'

interface Props {
  open: boolean
  childId: string
  type: 'weight' | 'height'
  editing?: Measurement | null
  onClose: () => void
  onSaved: () => void
}

const typeLabel: Record<string, string> = {
  weight: '体重 (kg)',
  height: '身高 (cm)',
}

const typeRange: Record<string, [number, number]> = {
  weight: [0.5, 200],
  height: [20, 250],
}

export default function MeasurementDrawer({
  open,
  childId,
  type,
  editing,
  onClose,
  onSaved,
}: Props) {
  const [form] = Form.useForm()

  useEffect(() => {
    if (!open) return
    if (editing) {
      form.setFieldsValue({
        value: editing.value,
        measured_at: dayjs(editing.measured_at),
        note: editing.note ?? '',
      })
    } else {
      form.resetFields()
      form.setFieldsValue({ measured_at: dayjs() })
    }
  }, [open, editing])

  const handleSave = async (values: {
    value: number
    measured_at: dayjs.Dayjs
    note: string
  }) => {
    const payload = {
      type,
      value: values.value,
      measured_at: values.measured_at.format('YYYY-MM-DD'),
      note: values.note?.trim() || null,
    }
    try {
      if (editing) {
        await updateMeasurement(childId, editing.id, payload)
      } else {
        await createMeasurement(childId, payload)
      }
      onSaved()
      onClose()
    } catch (err: any) {
      message.error(err.response?.data?.message ?? '保存失败')
    }
  }

  const [min, max] = typeRange[type] ?? [0, 9999]

  return (
    <Drawer
      title={editing ? `编辑${typeLabel[type]}` : `添加${typeLabel[type]}`}
      placement="right"
      open={open}
      onClose={onClose}
      width={360}
    >
      <Form form={form} layout="vertical" onFinish={handleSave}>
        <Form.Item
          name="value"
          label={typeLabel[type]}
          rules={[
            { required: true, message: '请输入数值' },
            {
              type: 'number',
              min,
              max,
              message: `数值应在 ${min}–${max} 之间`,
            },
          ]}
        >
          <InputNumber style={{ width: '100%' }} step={0.1} />
        </Form.Item>
        <Form.Item
          name="measured_at"
          label="测量日期"
          rules={[{ required: true }]}
        >
          <DatePicker style={{ width: '100%' }} />
        </Form.Item>
        <Form.Item name="note" label="备注（选填）">
          <Input.TextArea rows={3} />
        </Form.Item>
        <Button
          type="primary"
          htmlType="submit"
          block
          style={{ background: '#16a34a' }}
        >
          保存
        </Button>
      </Form>
    </Drawer>
  )
}
