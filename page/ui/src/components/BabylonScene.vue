<template>
  <canvas ref="canvas" class="babylon-canvas"></canvas>
</template>

<script lang="ts" setup>
import {onMounted, ref} from 'vue';
import * as BABYLON from '@babylonjs/core';

const canvas = ref<HTMLCanvasElement | null>(null);

onMounted(() => {
  if (!canvas.value) return;

  // 创建Babylon引擎
  const engine = new BABYLON.Engine(canvas.value, true, {}, true);

  // 创建场景
  const scene = new BABYLON.Scene(engine);

  // 添加相机
  const camera = new BABYLON.ArcRotateCamera("Camera",
      Math.PI / 2,
      Math.PI / 2,
      5,
      new BABYLON.Vector3(0, 0, 0));
  camera.attachControl(canvas.value, true);

  // 添加光源
  const light = new BABYLON.HemisphericLight("light", new BABYLON.Vector3(0, 1, 0), scene);
  light.diffuse = new BABYLON.Color3(1, 1, 1);

  // 添加服务器网格
  const server = BABYLON.MeshBuilder.CreateBox("server", {width: 2, height: 1.5, depth: 1}, scene);
  server.position = new BABYLON.Vector3(0, 0.75, 0); // 将服务器置于场景中心
  // 定义六种颜色，对应六个面
  const materials = [
    new BABYLON.StandardMaterial("frontColor", scene), // 前面
    new BABYLON.StandardMaterial("backColor", scene),  // 后面
    new BABYLON.StandardMaterial("topColor", scene),   // 顶面
    new BABYLON.StandardMaterial("bottomColor", scene), // 底面
    new BABYLON.StandardMaterial("rightColor", scene),  // 右面
    new BABYLON.StandardMaterial("leftColor", scene)   // 左面
  ];

  // 分配颜色到每个材质
  materials[0].diffuseColor = new BABYLON.Color3(1, 0, 0); // 前面红色
  materials[1].diffuseColor = new BABYLON.Color3(0, 1, 0); // 后面绿色
  materials[2].diffuseColor = new BABYLON.Color3(0, 0, 1); // 顶面蓝色
  materials[3].diffuseColor = new BABYLON.Color3(1, 1, 0); // 底面黄色
  materials[4].diffuseColor = new BABYLON.Color3(1, 0, 1); // 右面品红
  materials[5].diffuseColor = new BABYLON.Color3(0, 1, 1); // 左面青色

  // 创建多材质并分配
  const multiMaterial = new BABYLON.MultiMaterial("multiMat", scene);
  materials.forEach((mat, index) => {
    multiMaterial.subMaterials[index] = mat;
  });

  console.log(multiMaterial)

  // 应用多材质到服务器模型
  server.material = multiMaterial;


  // 渲染循环
  engine.runRenderLoop(() => {
    scene.render();
  });

  // 适应窗口大小变化
  window.addEventListener("resize", () => {
    engine.resize();
  });
});

</script>
<style scoped>
.babylon-canvas {
  width: 100%;
  height: 100%;
}
</style>