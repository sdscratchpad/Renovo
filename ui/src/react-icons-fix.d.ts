// Patch react-icons v5 IconType to be compatible with React 19 strict JSX type checking.
// react-icons v5 declares IconType as returning ReactNode (which includes undefined),
// but React 19 requires JSX component functions to return React.JSX.Element | null.
declare module 'react-icons/lib' {
  export type IconType = (props: IconBaseProps) => import('react').JSX.Element;
}
