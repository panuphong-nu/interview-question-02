<script setup lang="ts">
defineProps<{
  title: string
}>()
</script>

<template>
  <section class="panel">
    <aside class="panel__brand">
      <div class="brand-mark">
        <img
          class="brand-mark__image"
          src="https://companieslogo.com/img/orig/Y92.SI-112e04d0.png?download=true&t=1720244494"
          alt="ThaiBev"
          loading="eager"
          decoding="async"
        />
      </div>

      <div class="panel__brand-copy">
        <p class="panel__eyebrow panel__eyebrow--light">MEMBER PORTAL</p>
        <h2>เติบโตไปด้วยกัน<br />อย่างมั่นคงและยั่งยืน</h2>
        <p>ระบบสมาชิกที่เรียบง่าย ปลอดภัย และพร้อมใช้งานในทุกอุปกรณ์</p>
      </div>

      <div class="brand-accent" aria-hidden="true">
        <span></span>
        <span></span>
        <span></span>
      </div>
    </aside>

    <div class="panel__content">
      <header class="panel__header">
        <p class="panel__eyebrow">THAIBEV DIGITAL SERVICES</p>
        <h1 class="panel__title">{{ title }}</h1>
        <p class="panel__description">ระบบจัดการบัญชีผู้ใช้งาน</p>
      </header>

      <div class="panel__body">
        <slot />
      </div>

      <footer class="panel__footer">Thai Beverage Public Company Limited</footer>
    </div>
  </section>
</template>

<style scoped>
.panel {
  width: min(70rem, 100%);
  min-height: 38rem;
  display: grid;
  grid-template-columns: minmax(18rem, 0.78fr) minmax(25rem, 1.22fr);
  overflow: hidden;
  border: 1px solid rgb(0 74 46 / 12%);
  border-radius: 0.875rem;
  background: var(--color-surface);
  box-shadow: var(--shadow-panel);
}

.panel__brand {
  position: relative;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  min-height: 100%;
  padding: 3rem;
  overflow: hidden;
  color: #ffffff;
  background: var(--color-brand-dark);
}

.brand-mark {
  position: relative;
  z-index: 1;
  width: fit-content;
  padding: 0.85rem 1.1rem;
  border-left: 3px solid var(--color-gold);
  background: #ffffff;
}

.brand-mark__image {
  display: block;
  width: 9rem;
  height: auto;
}

.panel__brand-copy {
  position: relative;
  z-index: 1;
  max-width: 23rem;
  margin-block: 4rem;
}

.panel__brand-copy h2 {
  margin: 0.75rem 0 1rem;
  font-size: clamp(1.8rem, 3vw, 2.6rem);
  font-weight: 700;
  line-height: 1.35;
  letter-spacing: -0.025em;
}

.panel__brand-copy > p:last-child {
  margin: 0;
  color: rgb(255 255 255 / 74%);
  font-size: 0.95rem;
  line-height: 1.75;
}

.brand-accent {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: flex-end;
  gap: 0.45rem;
  height: 2.5rem;
}

.brand-accent span {
  display: block;
  width: 3.5rem;
  height: 0.2rem;
  background: var(--color-gold);
}

.brand-accent span:nth-child(2) {
  width: 1.5rem;
  background: rgb(255 255 255 / 40%);
}

.brand-accent span:nth-child(3) {
  width: 0.4rem;
  height: 0.4rem;
  border-radius: 50%;
}

.panel__brand::after {
  content: '';
  position: absolute;
  right: -10rem;
  bottom: -10rem;
  width: 25rem;
  height: 25rem;
  border: 1px solid rgb(200 168 90 / 32%);
  border-radius: 50%;
  box-shadow:
    0 0 0 3rem rgb(200 168 90 / 5%),
    0 0 0 6rem rgb(255 255 255 / 3%);
}

.panel__content {
  min-width: 0;
  display: flex;
  flex-direction: column;
  padding: 3.25rem clamp(2rem, 6vw, 4.5rem) 1.75rem;
}

.panel__header {
  padding-bottom: 1.75rem;
  border-bottom: 1px solid #e6ebe7;
}

.panel__eyebrow {
  margin: 0;
  color: var(--color-brand);
  font-size: 0.7rem;
  font-weight: 800;
  letter-spacing: 0.15em;
}

.panel__eyebrow--light {
  color: var(--color-gold-light);
}

.panel__title {
  margin: 0.45rem 0 0;
  color: var(--color-text);
  font-size: clamp(1.8rem, 4vw, 2.5rem);
  font-weight: 750;
  line-height: 1.2;
  letter-spacing: -0.04em;
}

.panel__description {
  margin: 0.5rem 0 0;
  color: var(--color-muted);
  font-size: 0.9rem;
}

.panel__body {
  flex: 1;
  display: flex;
  align-items: center;
  padding-block: 2rem;
}

.panel__body > * {
  width: 100%;
}

.panel__footer {
  padding-top: 1rem;
  border-top: 1px solid #e6ebe7;
  color: #879189;
  font-size: 0.68rem;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

@media (max-width: 52rem) {
  .panel {
    grid-template-columns: 1fr;
    min-height: auto;
  }

  .panel__brand {
    min-height: auto;
    padding: 1.5rem;
  }

  .brand-mark {
    padding: 0.65rem 0.85rem;
  }

  .brand-mark__image {
    width: 7.5rem;
  }

  .panel__brand-copy {
    margin: 2.5rem 0 1rem;
  }

  .panel__brand-copy h2 {
    font-size: 1.75rem;
  }

  .panel__brand-copy > p:last-child,
  .brand-accent {
    display: none;
  }

  .panel__content {
    padding: 2rem 1.5rem 1.25rem;
  }
}
</style>
