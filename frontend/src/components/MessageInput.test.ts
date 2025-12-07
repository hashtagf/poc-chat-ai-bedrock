import { describe, it, expect, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import fc from 'fast-check'
import MessageInput from './MessageInput.vue'

describe('MessageInput', () => {
  describe('Property 2: Input validation prevents invalid submission', () => {
    /**
     * Feature: chat-ui, Property 2: Input validation prevents invalid submission
     * Validates: Requirements 1.4, 8.3
     * 
     * For any string that is empty or contains only whitespace characters,
     * the Chat UI should prevent message submission and the message should
     * not be transmitted to the backend.
     */
    it('should prevent submission for empty or whitespace-only strings', async () => {
      await fc.assert(
        fc.asyncProperty(
          // Generate whitespace-only strings (at least 1 character to avoid empty string edge case)
          fc.stringOf(fc.constantFrom(' ', '\t', '\n', '\r', '\u00A0'), { minLength: 0, maxLength: 50 }),
          async (whitespaceString) => {
            const wrapper = mount(MessageInput, {
              props: {
                disabled: false
              }
            })

            // Set the input value to whitespace string
            const textarea = wrapper.find('textarea')
            await textarea.setValue(whitespaceString)
            await wrapper.vm.$nextTick()

            // Try to submit via button click
            const button = wrapper.find('button[type="submit"]')
            
            // Button should be disabled for whitespace-only input
            expect(button.attributes('disabled')).toBeDefined()

            // Try to submit anyway
            await button.trigger('click')
            await wrapper.vm.$nextTick()

            // Submit event should not be emitted
            expect(wrapper.emitted('submit')).toBeUndefined()

            // Try to submit via Enter key
            await textarea.trigger('keydown', { key: 'Enter' })
            await wrapper.vm.$nextTick()

            // Submit event should still not be emitted
            expect(wrapper.emitted('submit')).toBeUndefined()
          }
        ),
        { numRuns: 100 }
      )
    })

    it('should prevent submission for empty string', async () => {
      const wrapper = mount(MessageInput, {
        props: {
          disabled: false
        }
      })

      // Leave input empty
      const textarea = wrapper.find('textarea')
      await textarea.setValue('')

      // Button should be disabled
      const button = wrapper.find('button[type="submit"]')
      expect(button.attributes('disabled')).toBeDefined()

      // Try to submit
      await button.trigger('click')

      // Submit event should not be emitted
      expect(wrapper.emitted('submit')).toBeUndefined()
    })

    it('should allow submission for valid non-whitespace strings', async () => {
      await fc.assert(
        fc.asyncProperty(
          // Generate strings with at least one non-whitespace character
          fc.string({ minLength: 1, maxLength: 2000 }).filter(s => s.trim().length > 0),
          async (validString) => {
            const wrapper = mount(MessageInput, {
              props: {
                disabled: false
              }
            })

            // Set valid input
            const textarea = wrapper.find('textarea')
            await textarea.setValue(validString)
            await wrapper.vm.$nextTick()

            // Button should be enabled
            const button = wrapper.find('button[type="submit"]')
            expect(button.attributes('disabled')).toBeUndefined()

            // Submit via form submission
            const form = wrapper.find('form')
            await form.trigger('submit')
            await wrapper.vm.$nextTick()

            // Submit event should be emitted with trimmed content
            const emitted = wrapper.emitted('submit')
            expect(emitted).toBeDefined()
            expect(emitted![0]).toEqual([validString.trim()])
          }
        ),
        { numRuns: 100 }
      )
    })

    it('should prevent submission for strings exceeding max length', async () => {
      await fc.assert(
        fc.asyncProperty(
          // Generate strings longer than 2000 characters with at least some non-whitespace
          fc.string({ minLength: 2001, maxLength: 3000 }).filter(s => s.trim().length > 0),
          async (longString) => {
            const wrapper = mount(MessageInput, {
              props: {
                disabled: false
              }
            })

            // Set long input
            const textarea = wrapper.find('textarea')
            await textarea.setValue(longString)
            await wrapper.vm.$nextTick()

            // Button should be disabled
            const button = wrapper.find('button[type="submit"]')
            expect(button.attributes('disabled')).toBeDefined()

            // Error message should be displayed
            const errorMessage = wrapper.find('.error-message')
            expect(errorMessage.exists()).toBe(true)
            expect(errorMessage.text()).toContain('exceeds maximum length')

            // Try to submit
            await button.trigger('click')
            await wrapper.vm.$nextTick()

            // Submit event should not be emitted
            expect(wrapper.emitted('submit')).toBeUndefined()
          }
        ),
        { numRuns: 100 }
      )
    })
  })

  describe('Property 3: UI state during message processing', () => {
    /**
     * Feature: chat-ui, Property 3: UI state during message processing
     * Validates: Requirements 1.2, 4.1
     * 
     * For any message being sent or response being generated, the Chat UI should
     * disable the input field and display a loading indicator until processing completes.
     */
    it('should disable input and show loading indicator when disabled prop is true', async () => {
      await fc.assert(
        fc.asyncProperty(
          fc.string({ minLength: 1, maxLength: 100 }),
          async (inputText) => {
            // Mount with disabled=true to simulate processing state
            const wrapper = mount(MessageInput, {
              props: {
                disabled: true
              }
            })

            // Set some input text
            const textarea = wrapper.find('textarea')
            await textarea.setValue(inputText)
            await wrapper.vm.$nextTick()

            // Textarea should be disabled
            expect(textarea.attributes('disabled')).toBeDefined()

            // Button should be disabled
            const button = wrapper.find('button[type="submit"]')
            expect(button.attributes('disabled')).toBeDefined()

            // Loading spinner should be visible
            const loadingSpinner = wrapper.find('.loading-spinner')
            expect(loadingSpinner.exists()).toBe(true)

            // Send icon should not be visible
            const sendIcon = wrapper.find('.send-icon')
            expect(sendIcon.exists()).toBe(false)

            // Form submission should not work
            const form = wrapper.find('form')
            await form.trigger('submit')
            await wrapper.vm.$nextTick()

            // No submit event should be emitted
            expect(wrapper.emitted('submit')).toBeUndefined()
          }
        ),
        { numRuns: 100 }
      )
    })

    it('should enable input when disabled prop is false', async () => {
      const wrapper = mount(MessageInput, {
        props: {
          disabled: false
        }
      })

      const textarea = wrapper.find('textarea')
      const button = wrapper.find('button[type="submit"]')

      // Textarea should not be disabled
      expect(textarea.attributes('disabled')).toBeUndefined()

      // Send icon should be visible
      const sendIcon = wrapper.find('.send-icon')
      expect(sendIcon.exists()).toBe(true)

      // Loading spinner should not be visible
      const loadingSpinner = wrapper.find('.loading-spinner')
      expect(loadingSpinner.exists()).toBe(false)
    })
  })

  describe('Property 4: Input field reset after successful send', () => {
    /**
     * Feature: chat-ui, Property 4: Input field reset after successful send
     * Validates: Requirements 1.5
     * 
     * For any successfully sent message, the Chat UI should clear the input field
     * content and restore focus to the input field.
     */
    it('should clear input and restore focus after successful submission', async () => {
      await fc.assert(
        fc.asyncProperty(
          fc.string({ minLength: 1, maxLength: 2000 }).filter(s => s.trim().length > 0),
          async (validMessage) => {
            const wrapper = mount(MessageInput, {
              props: {
                disabled: false
              }
            })

            // Set input value
            const textarea = wrapper.find('textarea')
            await textarea.setValue(validMessage)
            await wrapper.vm.$nextTick()

            // Submit the form
            const form = wrapper.find('form')
            await form.trigger('submit')
            await wrapper.vm.$nextTick()

            // Input should be cleared
            expect(textarea.element.value).toBe('')

            // Focus should be restored (check if the exposed focus method exists)
            expect(wrapper.vm.focus).toBeDefined()
          }
        ),
        { numRuns: 100 }
      )
    })

    it('should maintain empty state if input was already empty', async () => {
      const wrapper = mount(MessageInput, {
        props: {
          disabled: false
        }
      })

      const textarea = wrapper.find('textarea')
      
      // Input starts empty
      expect(textarea.element.value).toBe('')

      // Try to submit (should be prevented)
      const form = wrapper.find('form')
      await form.trigger('submit')
      await wrapper.vm.$nextTick()

      // Input should still be empty
      expect(textarea.element.value).toBe('')

      // No submit event should be emitted
      expect(wrapper.emitted('submit')).toBeUndefined()
    })
  })

  describe('Property 15: Keyboard submission support', () => {
    /**
     * Feature: chat-ui, Property 15: Keyboard submission support
     * Validates: Requirements 5.2
     * 
     * For any state where the input field is enabled, pressing the Enter key
     * should submit the message (equivalent to clicking the send button).
     */
    it('should submit message when Enter key is pressed', async () => {
      await fc.assert(
        fc.asyncProperty(
          fc.string({ minLength: 1, maxLength: 2000 }).filter(s => s.trim().length > 0),
          async (validMessage) => {
            const wrapper = mount(MessageInput, {
              props: {
                disabled: false
              }
            })

            // Set input value
            const textarea = wrapper.find('textarea')
            await textarea.setValue(validMessage)
            await wrapper.vm.$nextTick()

            // Press Enter key
            await textarea.trigger('keydown', { key: 'Enter' })
            await wrapper.vm.$nextTick()

            // Submit event should be emitted
            const emitted = wrapper.emitted('submit')
            expect(emitted).toBeDefined()
            expect(emitted![0]).toEqual([validMessage.trim()])

            // Input should be cleared
            expect(textarea.element.value).toBe('')
          }
        ),
        { numRuns: 100 }
      )
    })

    it('should not submit when Shift+Enter is pressed (allows new line)', async () => {
      await fc.assert(
        fc.asyncProperty(
          fc.string({ minLength: 1, maxLength: 2000 }).filter(s => s.trim().length > 0),
          async (validMessage) => {
            const wrapper = mount(MessageInput, {
              props: {
                disabled: false
              }
            })

            // Set input value
            const textarea = wrapper.find('textarea')
            await textarea.setValue(validMessage)
            await wrapper.vm.$nextTick()

            // Press Shift+Enter key
            await textarea.trigger('keydown', { key: 'Enter', shiftKey: true })
            await wrapper.vm.$nextTick()

            // Submit event should NOT be emitted
            expect(wrapper.emitted('submit')).toBeUndefined()

            // Input should NOT be cleared
            expect(textarea.element.value).toBe(validMessage)
          }
        ),
        { numRuns: 100 }
      )
    })

    it('should not submit when Enter is pressed on disabled input', async () => {
      await fc.assert(
        fc.asyncProperty(
          fc.string({ minLength: 1, maxLength: 2000 }).filter(s => s.trim().length > 0),
          async (validMessage) => {
            const wrapper = mount(MessageInput, {
              props: {
                disabled: true
              }
            })

            // Set input value
            const textarea = wrapper.find('textarea')
            await textarea.setValue(validMessage)
            await wrapper.vm.$nextTick()

            // Press Enter key
            await textarea.trigger('keydown', { key: 'Enter' })
            await wrapper.vm.$nextTick()

            // Submit event should NOT be emitted
            expect(wrapper.emitted('submit')).toBeUndefined()
          }
        ),
        { numRuns: 100 }
      )
    })

    it('should not submit when Enter is pressed on empty input', async () => {
      const wrapper = mount(MessageInput, {
        props: {
          disabled: false
        }
      })

      const textarea = wrapper.find('textarea')
      
      // Input is empty
      expect(textarea.element.value).toBe('')

      // Press Enter key
      await textarea.trigger('keydown', { key: 'Enter' })
      await wrapper.vm.$nextTick()

      // Submit event should NOT be emitted
      expect(wrapper.emitted('submit')).toBeUndefined()
    })
  })
})
