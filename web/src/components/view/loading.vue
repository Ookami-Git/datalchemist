<template>
    <section><span class="loader-49"></span></section>
</template>

<style lang="scss" scoped>
$base: #263238;
$lite:  white;
$brand: #212529;
$size: 48px;

*, *:after, *:before {
      box-sizing: border-box;
    }
::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

/* Track */
::-webkit-scrollbar-track {
  background: $base; 
}
 
/* Handle */
::-webkit-scrollbar-thumb {
  background: $brand; 
}

/* Handle on hover */
::-webkit-scrollbar-thumb:hover {
  background: #fff; 
}

section {
    min-width: 200px;
    height: 80vh;
    padding: 10px;
    position: relative;
    display: flex;
    align-items: center;
    justify-content: center;
    color: #ccc;
    transition: 0.3s linear;

    @media (max-width: 768px ) {
        width: 50%;
    }
    @media (max-width: 480px) {
    width: 100%;
    }
}

// Loaders

  .loader-49 {
    width : $size;
    height: $size;
    display: inline-block;
    position: relative;
    border-radius: 5%;
    &::after , &::before {
      content: '';
      width : $size;
      height: $size;
      border:2px solid $lite;
      border-radius: 5%;
      position: absolute;
      left:0;
      top: 0;
      animation: rotation 2s ease-in-out infinite alternate;
    }
    &::after {
      border-color: $brand;
      animation-direction: alternate-reverse;
    }
  }

  /* keyFrames */ 
  
  @keyframes rotation {
    0% { transform: rotate(0deg) }
    100% { transform: rotate(360deg) }
  }

</style>