import{O as h,a as T,b as x}from"./OnboardingPage-220988ef.js";import{d as C,k as G,S as R,T as w,U as K,m as P,R as k,a as r,o as u,b as p,w as e,e as o,f as n,p as d,J as M,_ as O}from"./index-d50afca2.js";const U={class:"graph-list mb-6"},N={class:"radio-button-group"},A=C({__name:"ConfigurationTypes",setup(B){const i=G(),m=R(),_=w(),c={postgres:K(),memory:_,kubernetes:m},t=P(i("KUMA_STORE_TYPE")),g=k(()=>c[t.value]);return(z,a)=>{const v=r("RouteTitle"),l=r("KRadio"),b=r("AppView"),f=r("RouteView");return u(),p(f,{name:"onboarding-configuration-types"},{default:e(({can:y,t:V})=>[o(v,{title:V("onboarding.routes.configuration-types.title"),render:!1},null,8,["title"]),n(),o(b,null,{default:e(()=>[o(h,{"with-image":""},{header:e(()=>[o(T,null,{title:e(()=>[n(`
              Learn about configuration storage
            `)]),_:1})]),content:e(()=>[d("div",U,[(u(),p(M(g.value)))]),n(),d("div",N,[o(l,{modelValue:t.value,"onUpdate:modelValue":a[0]||(a[0]=s=>t.value=s),name:"deployment","selected-value":"kubernetes"},{default:e(()=>[n(`
              Kubernetes
            `)]),_:1},8,["modelValue"]),n(),o(l,{modelValue:t.value,"onUpdate:modelValue":a[1]||(a[1]=s=>t.value=s),name:"deployment","selected-value":"postgres"},{default:e(()=>[n(`
              Postgres
            `)]),_:1},8,["modelValue"]),n(),o(l,{modelValue:t.value,"onUpdate:modelValue":a[2]||(a[2]=s=>t.value=s),name:"deployment","selected-value":"memory"},{default:e(()=>[n(`
              Memory
            `)]),_:1},8,["modelValue"])])]),navigation:e(()=>[o(x,{"next-step":y("use zones")?"onboarding-multi-zone":"onboarding-create-mesh","previous-step":"onboarding-deployment-types"},null,8,["next-step"])]),_:2},1024)]),_:2},1024)]),_:1})}}});const H=O(A,[["__scopeId","data-v-7be26533"]]);export{H as default};
