import { nextTick, ref } from 'vue';
import { useDialog } from 'naive-ui';
import type { FormInst } from 'naive-ui';

export interface FormFieldError {
  field: string;
  message: string;
}

interface FocusableField {
  focus: () => void;
}

/**
 * 弹窗表单共享行为:
 * - 校验失败时收集字段错误,在表单顶部展示可聚焦的错误摘要,并支持跳转到对应字段;
 * - 脏表单关闭时弹出确认,避免误触丢失输入。
 */
export function useFormModal() {
  const dialog = useDialog();
  const fieldErrors = ref<FormFieldError[]>([]);
  const summaryRef = ref<HTMLElement | null>(null);
  const fieldEls = new Map<string, FocusableField>();

  function registerField(field: string, element: unknown): void {
    const candidate = element as FocusableField | null;
    if (candidate && typeof candidate.focus === 'function') fieldEls.set(field, candidate);
    else fieldEls.delete(field);
  }

  function setSummaryRef(element: unknown): void {
    summaryRef.value = element instanceof HTMLElement ? element : null;
  }

  function focusField(field: string): void {
    fieldEls.get(field)?.focus();
  }

  function focusSummary(): void {
    void nextTick(() => summaryRef.value?.focus());
  }

  /** 执行校验并运行提交动作;校验失败时聚焦错误摘要并返回 false。 */
  async function submit(formRef: FormInst | null, action: () => Promise<boolean>): Promise<boolean> {
    try {
      await formRef?.validate();
    } catch (result) {
      fieldErrors.value = collectErrors(result);
      focusSummary();
      return false;
    }
    fieldErrors.value = [];
    return action();
  }

  /** 表单脏检查:打开弹窗时调用 snapshot(),之后 dirty() 与快照比较。 */
  function track(read: () => unknown): { snapshot: () => void; dirty: () => boolean } {
    let snapshotValue = JSON.stringify(read());
    return {
      snapshot: () => {
        snapshotValue = JSON.stringify(read());
      },
      dirty: () => JSON.stringify(read()) !== snapshotValue,
    };
  }

  /** 弹窗关闭守卫:脏表单先确认,确认后真正关闭。 */
  function guardClose(dirty: boolean, apply: (show: boolean) => void, show: boolean): void {
    if (show || !dirty) {
      apply(show);
      return;
    }
    dialog.warning({
      title: '放弃修改?',
      content: '表单内容尚未保存,关闭后将丢失当前输入。',
      positiveText: '放弃修改',
      negativeText: '继续编辑',
      onPositiveClick: () => apply(false),
    });
  }

  return { fieldErrors, summaryRef, setSummaryRef, registerField, focusField, submit, track, guardClose };
}

function collectErrors(result: unknown): FormFieldError[] {
  if (result && typeof result === 'object') {
    const fields = (result as { fields?: Record<string, Array<{ message?: string }>> }).fields;
    if (fields) {
      return Object.entries(fields).flatMap(([field, list]) =>
        (list ?? []).map((item) => ({ field, message: item.message ?? '校验未通过' })),
      );
    }
  }
  // naive-ui 的 validate 拒绝值:每个 NFormItem 一组错误项组成的嵌套数组。
  if (Array.isArray(result)) {
    return result.flatMap((entry) => {
      const items = Array.isArray(entry) ? entry : [entry];
      return items.map((item) => ({
        field: String((item as { field?: string })?.field ?? ''),
        message: String((item as { message?: string })?.message ?? '校验未通过'),
      }));
    });
  }
  return [{ field: '', message: '表单校验未通过' }];
}
