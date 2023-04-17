import{d as y,aq as b,ar as V,as as h,i as x,r as C,a5 as G,j as d,o as p,k as i,w as t,a as s,e as o,g as m,a2 as k,u as r,ap as u,_ as M}from"./index-cef1317e.js";import{O as P,a as T,b as B}from"./OnboardingPage-94aacede.js";const N={class:"graph-list mb-6"},O={class:"radio-button-group"},w=y({__name:"ConfigurationTypes",setup(K){const g=b(),c=V(),_={postgres:h(),memory:c,kubernetes:g},l=x(),e=C("kubernetes");G(function(){e.value=l.getters["config/getConfigurationType"]});const f=d(()=>l.getters["config/getMulticlusterStatus"]?"onboarding-multi-zone":"onboarding-create-mesh"),v=d(()=>_[e.value]);return(U,a)=>(p(),i(B,{"with-image":""},{header:t(()=>[s(P,null,{title:t(()=>[o(`
          Learn about configuration storage
        `)]),_:1})]),content:t(()=>[m("div",N,[(p(),i(k(r(v))))]),o(),m("div",O,[s(r(u),{modelValue:e.value,"onUpdate:modelValue":a[0]||(a[0]=n=>e.value=n),name:"deployment","selected-value":"kubernetes"},{default:t(()=>[o(`
          Kubernetes
        `)]),_:1},8,["modelValue"]),o(),s(r(u),{modelValue:e.value,"onUpdate:modelValue":a[1]||(a[1]=n=>e.value=n),name:"deployment","selected-value":"postgres"},{default:t(()=>[o(`
          Postgres
        `)]),_:1},8,["modelValue"]),o(),s(r(u),{modelValue:e.value,"onUpdate:modelValue":a[2]||(a[2]=n=>e.value=n),name:"deployment","selected-value":"memory"},{default:t(()=>[o(`
          Memory
        `)]),_:1},8,["modelValue"])])]),navigation:t(()=>[s(T,{"next-step":r(f),"previous-step":"onboarding-deployment-types"},null,8,["next-step"])]),_:1}))}});const z=M(w,[["__scopeId","data-v-673391df"]]);export{z as default};
