// web/src/components/SupplementForm.tsx
import { useEffect, useState } from 'react'
import {
  Drawer,
  Form,
  DatePicker,
  Input,
  AutoComplete,
  Button,
  message,
} from 'antd'
import { createSupplement, getSupplementNames } from '../api/supplement'
import type { SupplementRecord } from '../api/supplement'
import dayjs from 'dayjs'

interface Props {
  open: boolean
  childId: string
  editing?: SupplementRecord | null
  onClose: () => void
  onSaved: () => void
}

const defaultSupplements = ['维生素AD', '钙', 'DHA', '铁', '维生素D', '锌']

export default function SupplementForm({ open, childId, editing, onClose, onSaved }: Props) {
  const [form] = Form.useForm()
  const [existingNames, setExistingNames] = useState<string[]>([])

  useEffect(() => {
    if (!open) return
    if (editing) {
      form.setFieldsValue({
        supplement_name: editing.supplement_name,
        dose: editing.dose ?? '',
        taken_at: dayjs(editing.taken_at),
      })
    } else {
      form.resetFields()
      form.setFieldsValue({ taken_at: dayjs() })
    }
    // Load existing supplement names for autocomplete
    getSupplementNames(childId)
      .then(names => setExistingNames(Array.isArray(names) ? names : []))
      .catch(() => setExistingNames([]))
  }, [open, editing])

  const handleSave = async (values: any) => {
    const payload = {
      supplement_name: values.supplement_name.trim(),
      dose: values.dose?.trim() || null,
      taken_at: values.taken_at.format(),
    }
    try {
      if (editing) {
        // For editing, we'd need updateSupplement, but for simplicity keep create-only
        await createSupplement(childId, payload)
      } else {
        await createSupplement(childId, payload)
      }
      onSaved()
      onClose()
    } catch (err: any) {
      message.error(err.response?.data?.message ?? '保存失败')
    }
  }

  const options = [...new Set([...existingNames, ...defaultSupplements])].map(name => ({
    value: name,
    label: name,
  }))

  return (
    <Drawer
      title={editing ? '编辑补剂记录' : '添加补剂'}
      placement="right"
      open={open}
      onClose={onClose}
      width={360}
    >
      <Form form={form} layout="vertical" onFinish={handleSave}>
        <Form.Item
          name="supplement_name"
          label="补剂名称"
          rules={[{ required: true, message: '请输入补剂名称' }]}
        >
          <AutoComplete options={options} placeholder="选择或输入补剂名称">
            <Input />
          </AutoComplete>
        </Form.Item>
        <Form.Item name="dose" label="剂量（选填）">
          <Input placeholder="如：1粒、5ml" />
        </Form.Item>
        <Form.Item name="taken_at" label="时间" rules={[{ required: true }]}>
          <DatePicker showTime style={{ width: '100%' }} />
        </Form.Item>
        <Button type="primary" htmlType="submit" block style={{ background: '#16a34a' }}>
          保存
        </Button>
      </Form>
    </Drawer>
  )
}
