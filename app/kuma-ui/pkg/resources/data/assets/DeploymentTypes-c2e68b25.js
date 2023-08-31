import{d as f,B as v,c as y,J as b,o as u,a as r,w as e,h as t,b as s,g as n,i as d,t as h,j as V,O as p}from"./index-f1b8ae6a.js";import{O as z,a as x,b as S}from"./OnboardingPage-d7f0da66.js";import{y as G,z as M,e as w,h as B,A as D,_ as O,f as C}from"./RouteView.vue_vue_type_script_setup_true_lang-4a32e1ca.js";import{_ as N}from"./RouteTitle.vue_vue_type_script_setup_true_lang-6484968f.js";import"./WarningIcon.vue_vue_type_script_setup_true_lang-14dd845b.js";const T={class:"graph-list mb-6"},k={class:"radio-button-group"},A=f({__name:"DeploymentTypes",setup(I){const m=G(),c={standalone:M(),"multi-zone":m},_=w(),{t:i}=B(),a=v("standalone"),g=y(()=>c[a.value]);return b(function(){a.value=_.getters["config/getMulticlusterStatus"]?"multi-zone":"standalone"}),($,o)=>(u(),r(O,null,{default:e(()=>[t(N,{title:s(i)("onboarding.routes.deployment-types.title")},null,8,["title"]),n(),t(D,null,{default:e(()=>[t(z,{"with-image":""},{header:e(()=>[t(x,null,{title:e(()=>[n(`
              Learn about deployments
            `)]),description:e(()=>[d("p",null,h(s(i)("common.product.name"))+" can be deployed in standalone or multi-zone mode.",1)]),_:1})]),content:e(()=>[d("div",T,[(u(),r(V(g.value)))]),n(),d("div",k,[t(s(p),{modelValue:a.value,"onUpdate:modelValue":o[0]||(o[0]=l=>a.value=l),name:"mode","selected-value":"standalone","data-testid":"onboarding-standalone-radio-button"},{default:e(()=>[n(`
              Standalone deployment
            `)]),_:1},8,["modelValue"]),n(),t(s(p),{modelValue:a.value,"onUpdate:modelValue":o[1]||(o[1]=l=>a.value=l),name:"mode","selected-value":"multi-zone","data-testid":"onboarding-multi-zone-radio-button"},{default:e(()=>[n(`
              Multi-zone deployment
            `)]),_:1},8,["modelValue"])])]),navigation:e(()=>[t(S,{"next-step":"onboarding-configuration-types","previous-step":"onboarding-welcome"})]),_:1})]),_:1})]),_:1}))}});const q=C(A,[["__scopeId","data-v-3da9e9d7"]]);export{q as default};
