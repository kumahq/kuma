import{d as g,I as v,J as y,L as f,M as h,t as V,f as z,o as u,g as p,w as e,h as n,i as s,C as x,l as a,m as d,D as C,n as D,N as r,A as G,_ as w,q as M}from"./index-cf0727dc.js";import{O as N,a as S,b as B}from"./OnboardingPage-46f5b350.js";const O={class:"graph-list mb-6"},T={class:"radio-button-group"},I=g({__name:"DeploymentTypes",setup(k){const m=v(),c={standalone:y(),"multi-zone":m},{t:i}=f(),_=h(),o=V(_("use zones")?"multi-zone":"standalone"),b=z(()=>c[o.value]);return(L,t)=>(u(),p(w,null,{default:e(()=>[n(x,{title:s(i)("onboarding.routes.deployment-types.title")},null,8,["title"]),a(),n(G,null,{default:e(()=>[n(N,{"with-image":""},{header:e(()=>[n(S,null,{title:e(()=>[a(`
              Learn about deployments
            `)]),description:e(()=>[d("p",null,C(s(i)("common.product.name"))+" can be deployed in standalone or multi-zone mode.",1)]),_:1})]),content:e(()=>[d("div",O,[(u(),p(D(b.value)))]),a(),d("div",T,[n(s(r),{modelValue:o.value,"onUpdate:modelValue":t[0]||(t[0]=l=>o.value=l),name:"mode","selected-value":"standalone","data-testid":"onboarding-standalone-radio-button"},{default:e(()=>[a(`
              Standalone deployment
            `)]),_:1},8,["modelValue"]),a(),n(s(r),{modelValue:o.value,"onUpdate:modelValue":t[1]||(t[1]=l=>o.value=l),name:"mode","selected-value":"multi-zone","data-testid":"onboarding-multi-zone-radio-button"},{default:e(()=>[a(`
              Multi-zone deployment
            `)]),_:1},8,["modelValue"])])]),navigation:e(()=>[n(B,{"next-step":"onboarding-configuration-types","previous-step":"onboarding-welcome"})]),_:1})]),_:1})]),_:1}))}});const q=M(I,[["__scopeId","data-v-ebbd0722"]]);export{q as default};
