import{d as V,O as h,P as C,Q as k,L as x,t as G,f as N,o as l,g as i,w as e,h as a,i as r,C as P,l as n,p as B,j as O,m as p,n as w,N as u,F as M,k as T,A as F,_ as K,q as U}from"./index-0ab7ff60.js";import{O as $,a as z,b as A}from"./OnboardingPage-a7afc08b.js";const I={class:"graph-list mb-6"},L={class:"radio-button-group"},j=V({__name:"ConfigurationTypes",setup(q){const m=h(),c=C(),_={postgres:k(),memory:c,kubernetes:m},{t:g}=x(),t=G("kubernetes"),f=d=>{t.value=d.store.type},v=N(()=>_[t.value]);return(d,o)=>(l(),i(K,null,{default:e(({can:y})=>[a(P,{title:r(g)("onboarding.routes.configuration-types.title")},null,8,["title"]),n(),a(F,null,{default:e(()=>[a($,{"with-image":""},{header:e(()=>[a(z,null,{title:e(()=>[n(`
              Learn about configuration storage
            `)]),_:1})]),content:e(()=>[a(B,{src:"/config",onChange:f},{default:e(({data:b})=>[typeof b<"u"?(l(),O(M,{key:0},[p("div",I,[(l(),i(w(v.value)))]),n(),p("div",L,[a(r(u),{modelValue:t.value,"onUpdate:modelValue":o[0]||(o[0]=s=>t.value=s),name:"deployment","selected-value":"kubernetes"},{default:e(()=>[n(`
                  Kubernetes
                `)]),_:1},8,["modelValue"]),n(),a(r(u),{modelValue:t.value,"onUpdate:modelValue":o[1]||(o[1]=s=>t.value=s),name:"deployment","selected-value":"postgres"},{default:e(()=>[n(`
                  Postgres
                `)]),_:1},8,["modelValue"]),n(),a(r(u),{modelValue:t.value,"onUpdate:modelValue":o[2]||(o[2]=s=>t.value=s),name:"deployment","selected-value":"memory"},{default:e(()=>[n(`
                  Memory
                `)]),_:1},8,["modelValue"])])],64)):T("",!0)]),_:1})]),navigation:e(()=>[a(A,{"next-step":y("use zones")?"onboarding-multi-zone":"onboarding-create-mesh","previous-step":"onboarding-deployment-types"},null,8,["next-step"])]),_:2},1024)]),_:2},1024)]),_:1}))}});const Q=U(j,[["__scopeId","data-v-5db0c53c"]]);export{Q as default};
