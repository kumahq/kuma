import{d as S,i as r,o as a,a as n,w as e,j as l,g as f,a1 as p,k as t,b as i,H as h,J as y,t as d,e as c,_ as C}from"./index-DJJJbhb4.js";const D={class:"stack"},B={class:"columns"},K=S({__name:"MeshExternalServiceDetailView",props:{data:{}},setup(k){const u=k;return(o,E)=>{const v=r("KTruncate"),_=r("KBadge"),V=r("KCard"),b=r("AppView"),g=r("RouteView"),w=r("DataSource");return a(),n(w,{src:"/me"},{default:e(({data:m})=>[m?(a(),n(g,{key:0,name:"mesh-external-service-detail-view",params:{mesh:"",service:"",page:1,size:m.pageSize,s:"",dataPlane:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:e(()=>[l(b,null,{default:e(()=>[f("div",D,[l(V,null,{default:e(()=>[f("div",B,[u.data.status.addresses.length>0?(a(),n(p,{key:0},{title:e(()=>[t(`
                  Addresses
                `)]),body:e(()=>[l(v,null,{default:e(()=>[(a(!0),i(h,null,y(u.data.status.addresses,s=>(a(),i("span",{key:s.hostname},d(s.hostname),1))),128))]),_:1})]),_:1})):c("",!0),t(),o.data.spec.match?(a(),n(p,{key:1,class:"port"},{title:e(()=>[t(`
                  Port
                `)]),body:e(()=>[(a(!0),i(h,null,y([o.data.spec.match],s=>(a(),n(_,{key:s.port,appearance:"info"},{default:e(()=>[t(d(s.port)+"/"+d(s.protocol),1)]),_:2},1024))),128))]),_:1})):c("",!0),t(),o.data.spec.match?(a(),n(p,{key:2,class:"tls"},{title:e(()=>[t(`
                  TLS
                `)]),body:e(()=>[l(_,{appearance:"neutral"},{default:e(()=>{var s;return[t(d((s=o.data.spec.tls)!=null&&s.enabled?"Enabled":"Disabled"),1)]}),_:1})]),_:1})):c("",!0),t(),typeof o.data.status.vip<"u"?(a(),n(p,{key:3,class:"ip"},{title:e(()=>[t(`
                  VIP
                `)]),body:e(()=>[t(d(o.data.status.vip.ip),1)]),_:1})):c("",!0)])]),_:1})])]),_:1})]),_:2},1032,["params"])):c("",!0)]),_:1})}}}),T=C(K,[["__scopeId","data-v-1ccd6fb8"]]);export{T as default};
