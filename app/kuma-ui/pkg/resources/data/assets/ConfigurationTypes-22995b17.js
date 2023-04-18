import{d as y,aq as b,ar as V,as as h,h as x,r as C,a5 as G,i as d,o as p,j as i,w as o,a as s,e as t,f as m,a2 as M,u as r,ap as u,J as k}from"./index-5f1fbf13.js";import{O as P,a as T,b as N}from"./OnboardingPage-503a3bf7.js";const O={class:"graph-list mb-6"},w={class:"radio-button-group"},B=y({__name:"ConfigurationTypes",setup(K){const g=b(),c=V(),_={postgres:h(),memory:c,kubernetes:g},l=x(),e=C("kubernetes");G(function(){e.value=l.getters["config/getConfigurationType"]});const f=d(()=>l.getters["config/getMulticlusterStatus"]?"onboarding-multi-zone":"onboarding-create-mesh"),v=d(()=>_[e.value]);return(U,a)=>(p(),i(N,{"with-image":""},{header:o(()=>[s(P,null,{title:o(()=>[t(`
          Learn about configuration storage
        `)]),_:1})]),content:o(()=>[m("div",O,[(p(),i(M(r(v))))]),t(),m("div",w,[s(r(u),{modelValue:e.value,"onUpdate:modelValue":a[0]||(a[0]=n=>e.value=n),name:"deployment","selected-value":"kubernetes"},{default:o(()=>[t(`
          Kubernetes
        `)]),_:1},8,["modelValue"]),t(),s(r(u),{modelValue:e.value,"onUpdate:modelValue":a[1]||(a[1]=n=>e.value=n),name:"deployment","selected-value":"postgres"},{default:o(()=>[t(`
          Postgres
        `)]),_:1},8,["modelValue"]),t(),s(r(u),{modelValue:e.value,"onUpdate:modelValue":a[2]||(a[2]=n=>e.value=n),name:"deployment","selected-value":"memory"},{default:o(()=>[t(`
          Memory
        `)]),_:1},8,["modelValue"])])]),navigation:o(()=>[s(T,{"next-step":r(f),"previous-step":"onboarding-deployment-types"},null,8,["next-step"])]),_:1}))}});const z=k(B,[["__scopeId","data-v-673391df"]]);export{z as default};
