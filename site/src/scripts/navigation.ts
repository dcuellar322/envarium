document.querySelectorAll<HTMLDetailsElement>('details.mobile-menu').forEach((menu) => {
  const toggle = menu.querySelector('summary')
  const close = () => { menu.open = false }

  document.addEventListener('keydown', (event) => {
    if (event.key !== 'Escape' || !menu.open) return
    close()
    toggle?.focus()
    event.stopPropagation()
  })
  document.addEventListener('pointerdown', (event) => {
    if (event.target instanceof Node && !menu.contains(event.target)) close()
  })
  menu.addEventListener('focusout', (event) => {
    if (!(event.relatedTarget instanceof Node) || !menu.contains(event.relatedTarget)) close()
  })
  window.matchMedia('(min-width: 981px)').addEventListener('change', close)
})
