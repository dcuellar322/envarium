import { fireEvent, render, screen } from '@testing-library/vue'
import { expect, test } from 'vitest'
import AppDialog from '../../src/app/components/AppDialog.vue'

test('labels the modal, handles Escape cancellation, and returns keyboard focus', async () => {
  const trigger = document.createElement('button')
  trigger.textContent = 'Open review'
  document.body.append(trigger)
  trigger.focus()
  const view = render(AppDialog, {
    props: { labelledby: 'review-title' },
    slots: { default: '<h2 id="review-title">Review</h2><button>Continue</button>' },
  })
  const dialog = screen.getByRole('dialog', { name: 'Review' })
  expect(dialog).toHaveAttribute('aria-modal', 'true')
  screen.getByRole('button', { name: 'Continue' }).focus()
  await fireEvent(dialog, new Event('cancel', { cancelable: true }))
  expect(view.emitted().close).toHaveLength(1)
  view.unmount()
  expect(trigger).toHaveFocus()
  trigger.remove()
})
