/// <reference types="vite/client" />

// Augment MUI to accept the custom 'code' typography variant used in theme.ts
declare module '@mui/material/styles' {
  interface TypographyVariants {
    code: {
      fontFamily?: string
      fontSize?: string | number
      fontWeight?: string | number
      lineHeight?: string | number
      letterSpacing?: string | number
    }
  }
  interface TypographyVariantsOptions {
    code?: {
      fontFamily?: string
      fontSize?: string | number
      fontWeight?: string | number
      lineHeight?: string | number
      letterSpacing?: string | number
    }
  }
}

declare module '@mui/material/Typography' {
  interface TypographyPropsVariantOverrides {
    code: true
  }
}
