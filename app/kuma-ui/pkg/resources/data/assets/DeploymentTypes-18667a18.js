import{d as g,y as f,l as y,c as v,o as u,a as r,w as e,h as n,b as s,g as a,i as d,t as h,j as z,z as p}from"./index-9a3d231d.js";import{O as V,a as x,b as G}from"./OnboardingPage-e007feea.js";import{j as w,k as C,g as D,A as S,_ as B,f as M}from"./RouteView.vue_vue_type_script_setup_true_lang-da83f5a8.js";import{_ as N}from"./RouteTitle.vue_vue_type_script_setup_true_lang-3a51c48f.js";import"./WarningIcon.vue_vue_type_script_setup_true_lang-fe937ad6.js";const O={class:"graph-list mb-6"},T={class:"radio-button-group"},k=g({__name:"DeploymentTypes",setup(j){const m=w(),c={standalone:C(),"multi-zone":m},{t:i}=D(),_=f(),o=y(_("use zones")?"multi-zone":"standalone"),b=v(()=>c[o.value]);return(I,t)=>(u(),r(B,null,{default:e(()=>[n(N,{title:s(i)("onboarding.routes.deployment-types.title")},null,8,["title"]),a(),n(S,null,{default:e(()=>[n(V,{"with-image":""},{header:e(()=>[n(x,null,{title:e(()=>[a(`
              Learn about deployments
            `)]),description:e(()=>[d("p",null,h(s(i)("common.product.name"))+" can be deployed in standalone or multi-zone mode.",1)]),_:1})]),content:e(()=>[d("div",O,[(u(),r(z(b.value)))]),a(),d("div",T,[n(s(p),{modelValue:o.value,"onUpdate:modelValue":t[0]||(t[0]=l=>o.value=l),name:"mode","selected-value":"standalone","data-testid":"onboarding-standalone-radio-button"},{default:e(()=>[a(`
              Standalone deployment
            `)]),_:1},8,["modelValue"]),a(),n(s(p),{modelValue:o.value,"onUpdate:modelValue":t[1]||(t[1]=l=>o.value=l),name:"mode","selected-value":"multi-zone","data-testid":"onboarding-multi-zone-radio-button"},{default:e(()=>[a(`
              Multi-zone deployment
            `)]),_:1},8,["modelValue"])])]),navigation:e(()=>[n(G,{"next-step":"onboarding-configuration-types","previous-step":"onboarding-welcome"})]),_:1})]),_:1})]),_:1}))}});const q=M(k,[["__scopeId","data-v-ebbd0722"]]);export{q as default};
