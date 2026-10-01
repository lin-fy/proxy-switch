import type { GlobalThemeOverrides } from 'naive-ui';

/**
 * Naive UI 主题只通过公开 API 设置,颜色全部来自 tokens.css 的计算值。
 * 必须在 tokens.css 已生效后调用(在 App.vue setup 中调用即可),
 * 颜色运行时计算不接受 var() 表达式,所以这里读取实际值而不是直接传变量。
 */
function readToken(name: string, fallback: string): string {
  if (typeof window === 'undefined') return fallback;
  const value = getComputedStyle(document.documentElement).getPropertyValue(name).trim();
  return value || fallback;
}

export function createThemeOverrides(): GlobalThemeOverrides {
  const surface0 = readToken('--surface-0', '#0e1519');
  const surface1 = readToken('--surface-1', '#141e23');
  const surface2 = readToken('--surface-2', '#1a272b');
  const textPrimary = readToken('--text-primary', '#e7f0ee');
  const textSecondary = readToken('--text-secondary', '#a4b6b5');
  const textMuted = readToken('--text-muted', '#8ca3a1');
  const borderSubtle = readToken('--border-subtle', '#26363a');
  const accent = readToken('--accent', '#62e6c2');
  const accentStrong = readToken('--accent-strong', '#2ebb96');
  const accentPressed = readToken('--accent-pressed', '#189b7a');
  const warning = readToken('--warning', '#f0b45c');
  const warningStrong = readToken('--warning-strong', '#f5c47c');
  const danger = readToken('--danger', '#e77b80');
  const dangerStrong = readToken('--danger-strong', '#d95c66');
  const success = readToken('--success', '#62e6c2');
  const fontUi = readToken('--font-ui', "'Segoe UI', 'Microsoft YaHei UI', system-ui, sans-serif");
  const fontMono = readToken('--font-mono', "Consolas, 'Cascadia Mono', ui-monospace, monospace");

  return {
    common: {
      primaryColor: accent,
      primaryColorHover: accentStrong,
      primaryColorPressed: accentPressed,
      primaryColorSuppl: accentStrong,
      infoColor: accent,
      infoColorHover: accentStrong,
      infoColorPressed: accentPressed,
      infoColorSuppl: accentStrong,
      successColor: success,
      successColorHover: accentStrong,
      successColorPressed: accentPressed,
      successColorSuppl: accentStrong,
      warningColor: warning,
      warningColorHover: warningStrong,
      warningColorPressed: warning,
      warningColorSuppl: warningStrong,
      errorColor: danger,
      errorColorHover: dangerStrong,
      errorColorPressed: dangerStrong,
      errorColorSuppl: dangerStrong,
      bodyColor: surface0,
      cardColor: surface1,
      modalColor: surface2,
      popoverColor: surface2,
      tableColor: surface1,
      inputColor: surface2,
      actionColor: surface1,
      textColorBase: textPrimary,
      textColor1: textPrimary,
      textColor2: textSecondary,
      textColor3: textMuted,
      placeholderColor: textMuted,
      borderColor: borderSubtle,
      dividerColor: borderSubtle,
      borderRadius: '6px',
      borderRadiusSmall: '4px',
      fontFamily: fontUi,
      fontFamilyMono: fontMono,
      fontSizeMini: '11px',
      fontSizeTiny: '11px',
      fontSizeSmall: '12px',
      fontSizeMedium: '13px',
      fontSizeLarge: '14px',
      fontSizeHuge: '15px',
    },
    Button: {
      borderRadiusMedium: '6px',
      borderRadiusSmall: '5px',
      heightMedium: '32px',
      heightSmall: '28px',
      fontWeight: '600',
    },
    Input: {
      borderRadius: '6px',
      heightMedium: '32px',
    },
    Select: {
      peers: {
        InternalSelection: {
          borderRadius: '6px',
          heightMedium: '32px',
        },
      },
    },
    Switch: {
      railColorActive: accentStrong,
    },
    Skeleton: {
      color: surface2,
      colorEnd: surface1,
    },
    Modal: {
      borderRadius: '12px',
    },
    Dialog: {
      borderRadius: '10px',
    },
    Card: {
      borderRadius: '10px',
      paddingMedium: '16px 20px',
      color: surface1,
    },
    Tag: {
      borderRadius: '5px',
    },
    Empty: {
      textColor: textMuted,
      iconColor: textMuted,
    },
    Dropdown: {
      borderRadius: '8px',
    },
  };
}
