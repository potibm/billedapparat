/* eslint-disable @typescript-eslint/no-empty-object-type */
/* eslint-disable @typescript-eslint/no-explicit-any */
import { type TestingLibraryMatchers } from "@testing-library/jest-dom/types/matchers";

declare module "vitest" {
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  interface Assertion<R = void, T = any> extends TestingLibraryMatchers<
    any,
    R
  > {}
  interface AsymmetricMatchersContaining extends TestingLibraryMatchers<
    any,
    any
  > {}
}
