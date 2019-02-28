<template>
  <div>
    <Header/>
    <h2>What is Cheesecake Live?</h2>
    <p>Cheesecake is a prediction model and analytics program for the <em>FIRST</em> Robotics Competition.</p>
    <p>Cheesecake takes an empirical, evidence-based approach to how it handles FRC data.</p>
  </div>
</template>

<script>
import Header from '@/components/Header'

export default {
  name: 'Home',
  components: { Header }
}
</script>
