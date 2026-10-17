// Pinned native SVG renderers can differ by one intensity level on one channel
// across platforms. Every other pixel must still match exactly.
export function equalNativePixels(left, right) {
  if (left.length !== right.length) return false
  let differences = 0
  for (let index = 0; index < left.length; index++) {
    const delta = Math.abs(left[index] - right[index])
    if (delta === 0) continue
    if (delta > 1 || ++differences > 1) return false
  }
  return true
}
