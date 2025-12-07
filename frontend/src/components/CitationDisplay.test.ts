import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import fc from 'fast-check'
import CitationDisplay from './CitationDisplay.vue'
import type { Citation } from '../types'

// Generators for property-based testing

// Generate non-whitespace strings
const nonWhitespaceString = (minLength: number, maxLength: number) =>
  fc.string({ minLength, maxLength })
    .filter(s => s.trim().length > 0)

const citationArb = fc.record({
  sourceId: fc.uuid(),
  sourceName: nonWhitespaceString(1, 100),
  excerpt: nonWhitespaceString(10, 500),
  confidence: fc.option(fc.float({ min: 0, max: 1 }), { nil: undefined }),
  url: fc.option(fc.webUrl(), { nil: undefined }),
  metadata: fc.option(fc.dictionary(
    nonWhitespaceString(1, 20),
    fc.oneof(fc.string(), fc.integer(), fc.boolean())
  ), { nil: undefined })
})

// Generate array of citations with unique sourceIds
const citationsArrayArb = fc.array(citationArb, { minLength: 1, maxLength: 10 })
  .map(citations => {
    // Ensure unique sourceIds by regenerating them
    return citations.map((citation, index) => ({
      ...citation,
      sourceId: `${citation.sourceId}-${index}`
    }))
  })

describe('CitationDisplay', () => {
  describe('Property 28: Citation interaction reveals details', () => {
    // Feature: chat-ui, Property 28: Citation interaction reveals details
    // Validates: Requirements 9.2
    it('should reveal source information when user interacts with citation indicator', async () => {
      await fc.assert(
        fc.asyncProperty(citationsArrayArb, async (citations) => {
          const wrapper = mount(CitationDisplay, {
            props: { citations }
          })

          // For each citation, verify interaction reveals details
          for (let index = 0; index < citations.length; index++) {
            const citation = citations[index]
            
            // Initially, details should not be visible
            const detailsId = `citation-details-${citation.sourceId}`
            let detailsElement = wrapper.find(`#${detailsId}`)
            expect(detailsElement.exists()).toBe(false)

            // Find and click the citation header button
            const buttons = wrapper.findAll('button.citation-header')
            expect(buttons.length).toBeGreaterThan(index)
            
            const button = buttons[index]
            await button.trigger('click')

            // After click, details should be visible
            detailsElement = wrapper.find(`#${detailsId}`)
            expect(detailsElement.exists()).toBe(true)

            // Verify the details contain the required information
            const detailsText = detailsElement.text()
            
            // Should contain source name
            expect(wrapper.text()).toContain(citation.sourceName)
            
            // Should contain excerpt (normalize whitespace since HTML collapses it)
            // Trim both strings after normalizing to handle leading/trailing whitespace
            const normalizedDetailsText = detailsText.replace(/\s+/g, ' ').trim()
            const normalizedExcerpt = citation.excerpt.replace(/\s+/g, ' ').trim()
            expect(normalizedDetailsText).toContain(normalizedExcerpt)
            
            // Should contain confidence if available
            if (citation.confidence !== undefined) {
              const confidencePercent = Math.round(citation.confidence * 100)
              expect(wrapper.text()).toContain(`${confidencePercent}%`)
            }
          }
        }),
        { numRuns: 100 }
      )
    })
  })

  describe('Property 29: Multi-citation distinction', () => {
    // Feature: chat-ui, Property 29: Multi-citation distinction
    // Validates: Requirements 9.3
    it('should distinguish between different sources clearly when multiple citations exist', () => {
      fc.assert(
        fc.property(
          fc.array(citationArb, { minLength: 2, maxLength: 10 }),
          (citations) => {
            const wrapper = mount(CitationDisplay, {
              props: { citations }
            })

            // Verify each citation has a unique badge number (provides distinction)
            const badges = wrapper.findAll('.citation-badge')
            expect(badges.length).toBe(citations.length)

            badges.forEach((badge, index) => {
              expect(badge.text()).toBe(`${index + 1}`)
            })

            // Verify each citation has the citation-multiple class
            const citationItems = wrapper.findAll('.citation-item')
            citationItems.forEach((item) => {
              expect(item.classes()).toContain('citation-multiple')
            })

            // Verify each citation displays its unique source name
            const sourceNames = wrapper.findAll('.citation-source-name')
            expect(sourceNames.length).toBe(citations.length)
            
            citations.forEach((citation, index) => {
              // HTML rendering trims whitespace, so we compare trimmed values
              expect(sourceNames[index].text().trim()).toBe(citation.sourceName.trim())
            })

            // Verify visual distinction exists through CSS classes
            // Each citation should have distinct styling classes
            const firstFourItems = citationItems.slice(0, Math.min(4, citations.length))
            
            firstFourItems.forEach((item) => {
              const classes = item.classes()
              // Each item should have the citation-multiple class which applies distinct styling
              expect(classes).toContain('citation-multiple')
              // The component uses nth-child CSS selectors for visual distinction
              // We verify the structure is correct for CSS to apply
            })
          }
        ),
        { numRuns: 100 }
      )
    })
  })

  describe('Property 30: Non-cited response indication', () => {
    // Feature: chat-ui, Property 30: Non-cited response indication
    // Validates: Requirements 9.4
    it('should indicate response is based on general knowledge when no citations provided', () => {
      fc.assert(
        fc.property(
          fc.constantFrom(undefined, [], null),
          (noCitations) => {
            const wrapper = mount(CitationDisplay, {
              props: { citations: noCitations as any }
            })

            // Should display the no-citations indicator
            const noCitationsElement = wrapper.find('.no-citations')
            expect(noCitationsElement.exists()).toBe(true)

            // Should contain text indicating general knowledge
            const text = noCitationsElement.text()
            expect(text.toLowerCase()).toContain('general knowledge')

            // Should not display citations list
            const citationsList = wrapper.find('.citations-list')
            expect(citationsList.exists()).toBe(false)

            // Should have appropriate ARIA label
            const displayElement = wrapper.find('.citation-display')
            const ariaLabel = displayElement.attributes('aria-label')
            expect(ariaLabel).toContain('No citations')
          }
        ),
        { numRuns: 100 }
      )
    })
  })

  describe('Property 31: Citation metadata display', () => {
    // Feature: chat-ui, Property 31: Citation metadata display
    // Validates: Requirements 9.5
    it('should display confidence scores and metadata alongside citations when available', async () => {
      await fc.assert(
        fc.asyncProperty(
          fc.array(
            fc.record({
              sourceId: fc.uuid(),
              sourceName: nonWhitespaceString(1, 100),
              excerpt: nonWhitespaceString(10, 500),
              confidence: fc.float({ min: 0, max: 1 }),
              url: fc.option(fc.webUrl(), { nil: undefined }),
              metadata: fc.dictionary(
                nonWhitespaceString(1, 20),
                fc.oneof(nonWhitespaceString(1, 50), fc.integer(), fc.boolean()),
                { minKeys: 1, maxKeys: 5 }
              )
            }),
            { minLength: 1, maxLength: 5 }
          ).map(citations => {
            // Ensure unique sourceIds
            return citations.map((citation, index) => ({
              ...citation,
              sourceId: `${citation.sourceId}-${index}`
            }))
          }),
          async (citations) => {
            const wrapper = mount(CitationDisplay, {
              props: { citations }
            })

            for (let index = 0; index < citations.length; index++) {
              const citation = citations[index]
              
              // Verify confidence score is displayed
              const confidencePercent = Math.round(citation.confidence * 100)
              const confidenceElements = wrapper.findAll('.citation-confidence')
              expect(confidenceElements.length).toBeGreaterThan(index)
              expect(confidenceElements[index].text()).toBe(`${confidencePercent}%`)

              // Expand the citation to see metadata
              const buttons = wrapper.findAll('button.citation-header')
              await buttons[index].trigger('click')

              // Verify metadata is displayed
              const detailsElement = wrapper.find(`#citation-details-${citation.sourceId}`)
              expect(detailsElement.exists()).toBe(true)

              if (citation.metadata && Object.keys(citation.metadata).length > 0) {
                const metadataSection = detailsElement.find('.citation-metadata')
                expect(metadataSection.exists()).toBe(true)

                // Verify each metadata key-value pair is displayed
                Object.entries(citation.metadata).forEach(([key, value]) => {
                  const metadataText = metadataSection.text()
                  // HTML rendering trims whitespace, so we check for trimmed values
                  expect(metadataText).toContain(key.trim())
                  expect(metadataText).toContain(String(value).trim())
                })
              }
            }
          }
        ),
        { numRuns: 100 }
      )
    })
  })

  // Additional unit tests for edge cases
  describe('Unit tests', () => {
    it('should render with no citations', () => {
      const wrapper = mount(CitationDisplay, {
        props: { citations: [] }
      })

      expect(wrapper.find('.no-citations').exists()).toBe(true)
      expect(wrapper.find('.citations-list').exists()).toBe(false)
    })

    it('should render with single citation', () => {
      const citation: Citation = {
        sourceId: 'test-1',
        sourceName: 'Test Source',
        excerpt: 'This is a test excerpt',
        confidence: 0.95
      }

      const wrapper = mount(CitationDisplay, {
        props: { citations: [citation] }
      })

      expect(wrapper.find('.no-citations').exists()).toBe(false)
      expect(wrapper.find('.citations-list').exists()).toBe(true)
      expect(wrapper.findAll('.citation-item').length).toBe(1)
    })

    it('should toggle citation details on click', async () => {
      const citation: Citation = {
        sourceId: 'test-1',
        sourceName: 'Test Source',
        excerpt: 'This is a test excerpt'
      }

      const wrapper = mount(CitationDisplay, {
        props: { citations: [citation] }
      })

      const button = wrapper.find('button.citation-header')
      
      // Initially collapsed
      expect(wrapper.find('#citation-details-test-1').exists()).toBe(false)

      // Click to expand
      await button.trigger('click')
      expect(wrapper.find('#citation-details-test-1').exists()).toBe(true)

      // Click to collapse
      await button.trigger('click')
      expect(wrapper.find('#citation-details-test-1').exists()).toBe(false)
    })

    it('should display URL link when available', async () => {
      const citation: Citation = {
        sourceId: 'test-1',
        sourceName: 'Test Source',
        excerpt: 'This is a test excerpt',
        url: 'https://example.com'
      }

      const wrapper = mount(CitationDisplay, {
        props: { citations: [citation] }
      })

      // Expand citation
      await wrapper.find('button.citation-header').trigger('click')

      const link = wrapper.find('.citation-link')
      expect(link.exists()).toBe(true)
      expect(link.attributes('href')).toBe('https://example.com')
      expect(link.attributes('target')).toBe('_blank')
      expect(link.attributes('rel')).toBe('noopener noreferrer')
    })

    it('should format confidence score as percentage', () => {
      const citation: Citation = {
        sourceId: 'test-1',
        sourceName: 'Test Source',
        excerpt: 'This is a test excerpt',
        confidence: 0.856
      }

      const wrapper = mount(CitationDisplay, {
        props: { citations: [citation] }
      })

      const confidence = wrapper.find('.citation-confidence')
      expect(confidence.text()).toBe('86%')
    })

    it('should expand all citations when expanded prop is true', () => {
      const citations: Citation[] = [
        {
          sourceId: 'test-1',
          sourceName: 'Test Source 1',
          excerpt: 'Excerpt 1'
        },
        {
          sourceId: 'test-2',
          sourceName: 'Test Source 2',
          excerpt: 'Excerpt 2'
        }
      ]

      const wrapper = mount(CitationDisplay, {
        props: { citations, expanded: true }
      })

      // Both citations should be expanded
      expect(wrapper.find('#citation-details-test-1').exists()).toBe(true)
      expect(wrapper.find('#citation-details-test-2').exists()).toBe(true)
    })
  })
})
