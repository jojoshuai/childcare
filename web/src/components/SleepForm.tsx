// web/src/components/SleepForm.tsx
import { useEffect } from 'react'
import {
  Drawer,
  Form,
  DatePicker,
  InputNumber,
  Checkbox,
  Button,
  message,
} from 'antd'
import { createSleep, updateSleep } from '../api/sleep'
import type { SleepRecord } from '../api/sleep'
import dayjs from 'dayjs'

interface Props {
  open: boolean
  childId: string
  editing?: SleepRecord | null
  onClose: () => void
  onSaved: () => void
}

export default function SleepForm({ open, childId, editing, onClose, onSaved }: Props) {
  const [form] = Form.useForm()

  useEffect(() => {
    if (!open) return
    if (editing) {
      form.setFieldsValue({
        start_time: dayjs(editing.start_time),
        end_time: editing.end_time ? dayjs(editing.end_time) : null,
        woke_up: editing.woke_up,
        wake_count: editing.wake_count,
      })
    } else {
      form.resetFields()
      form.setFieldsValue({ start_time: dayjs(), woke_up: false, wake_count: 0 })
    }
  }, [open, editing])

  const handleSave = async (values: any) => {
    const payload: any = {
      start_time: values.start_time.format(),
      end_time: values.end_time ? values.end_time.format() : undefined,
      woke_up: !!values.woke_up,
      wake_count: values.wake_count || 0,
    }
    try {
      if (editing) {
        await updateSleep(childId, editing.id, payload)
      } else {
        await createSleep(childId, payload)
      }
      onSaved()
      onClose()
    } catch (err: any) {
      message.error(err.response?.data?.message ?? '保存失败')
    }
  }

  const startVal = Form.useWatch('start_time', form)
  const endVal = Form.useWatch('end_time', form)
  let duration = ''
  if (startVal && endVal) {
    const diff = dayjs(endVal).diff(dayjs(startVal), 'minute')
    if (diff > 0) {
      const h = Math.floor(diff / 60)
      const m = diff % 60
      duration = `时长：${h > 0 ? `${h}小时` : ''}${m}分钟`
    }
  }

  return (
    <Drawer
      title={editing ? '编辑睡眠记录' : '添加睡眠记录'}
      placement="right"
      open={open}
      onClose={onClose}
      width={360}
    >
      <Form form={form} layout="vertical" onFinish={handleSave}>
        <Form.Item name="start_time" label="开始时间" rules={[{ required: true }]}>
          <DatePicker showTime style={{ width: '100%' }} />
        </Form.Item>
        <Form.Item name="end_time" label="结束时间（选填）">
          <DatePicker showTime style={{ width: '100%' }} />
        </Form.Item>
        {duration && (
          <div style={{ color: '#16a34a', marginBottom: 16, fontSize: 13 }}>{duration}</div>
        )}
        <Form.Item name="woke_up" valuePropName="checked">
          <Checkbox>中途醒来</Checkbox>
        </Form.Item>
        <Form.Item name="wake_count" label="醒来次数">
          <InputNumber min={0} max={20} style={{ width: '100%' }} />
        </Form.Item>
        <Button type="primary" htmlType="submit" block style={{ background: '#16a34a' }}>
          保存
        </Button>
      </Form>
    </Drawer>
  )
}
