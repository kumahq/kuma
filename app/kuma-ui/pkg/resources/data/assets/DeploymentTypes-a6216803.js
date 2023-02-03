import{m as g}from"./vuex.esm-bundler-4e6e06ec.js";import{f}from"./kongponents.es-7ead79da.js";import{P as b}from"./constants-31fdaf55.js";import{O as y}from"./OnboardingNavigation-14ec4750.js";import{O as h,a as v}from"./OnboardingPage-ce1119d2.js";import{d as O,e as x}from"./index-e2f1942d.js";import{_ as w}from"./_plugin-vue_export-helper-c27b6911.js";import{i as n,a as s,w as o,o as m,e as a,g as l,y as G,f as i}from"./runtime-dom.esm-bundler-fd3ecc5a.js";import"./production-8efaeab1.js";import"./index-be4d4b11.js";import"./DoughnutChart-210a9e41.js";import"./vue-router-67937a96.js";import"./store-ec4aec64.js";const N={name:"DeploymentTypes",components:{MultizoneGraph:O(),StandaloneGraph:x(),OnboardingNavigation:y,OnboardingHeading:h,OnboardingPage:v,KRadio:f},data(){return{mode:"standalone",productName:b}},computed:{...g({multicluster:"config/getMulticlusterStatus"}),currentGraph(){return this.mode==="standalone"?"StandaloneGraph":"MultizoneGraph"}},mounted(){this.mode=this.multicluster?"multi-zone":"standalone"}},V={class:"h-full w-full flex items-center justify-center mb-10"},z={class:"radio flex text-base justify-between w-full sm:w-3/4 md:w-3/5 lg:w-1/2 absolute bottom-0 right-0 left-0 mb-10 mx-auto deployment-type-radio-buttons"};function P(M,t,D,S,e,p){const u=n("OnboardingHeading"),r=n("KRadio"),c=n("OnboardingNavigation"),_=n("OnboardingPage");return m(),s(_,{"with-image":""},{header:o(()=>[a(u,{title:"Learn about deployments",description:`${e.productName} can be deployed in standalone or multi-zone mode.`},null,8,["description"])]),content:o(()=>[l("div",V,[(m(),s(G(p.currentGraph)))]),i(),l("div",z,[a(r,{modelValue:e.mode,"onUpdate:modelValue":t[0]||(t[0]=d=>e.mode=d),name:"mode","selected-value":"standalone","data-testid":"onboarding-standalone-radio-button"},{default:o(()=>[i(`
          Standalone deployment
        `)]),_:1},8,["modelValue"]),i(),a(r,{modelValue:e.mode,"onUpdate:modelValue":t[1]||(t[1]=d=>e.mode=d),name:"mode","selected-value":"multi-zone","data-testid":"onboarding-multi-zone-radio-button"},{default:o(()=>[i(`
          Multi-zone deployment
        `)]),_:1},8,["modelValue"])])]),navigation:o(()=>[a(c,{"next-step":"onboarding-configuration-types","previous-step":"onboarding-welcome"})]),_:1})}const q=w(N,[["render",P],["__scopeId","data-v-1868c73f"]]);export{q as default};
