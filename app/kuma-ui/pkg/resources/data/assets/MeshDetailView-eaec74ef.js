import{d as k,L as y,f as h,o as t,j as x,l as d,D as n,i as s,h as e,H as T,q as $,aa as V,g as r,w as _,C as g,A as B,p as D,E as w,s as M,m as R,_ as S}from"./index-18fd9432.js";const C={class:"date-status"},N=k({__name:"ResourceDateStatus",props:{creationTime:{},modificationTime:{}},setup(u){const i=u,{t:o,formatIsoDate:l}=y(),p=h(()=>l(i.creationTime)),c=h(()=>l(i.modificationTime));return(a,m)=>(t(),x("span",C,[d(n(s(o)("common.detail.created"))+": "+n(p.value)+" ",1),e(s(T),{icon:"arrowRight"}),d(" "+n(s(o)("common.detail.modified"))+": "+n(c.value),1)]))}});const b=$(N,[["__scopeId","data-v-5d50f5d4"]]),A={key:3,class:"stack","data-testid":"detail-view-details"},I={class:"date-status-wrapper"},j=k({__name:"MeshDetailView",setup(u){const{t:i}=y(),o=V();return(l,p)=>(t(),r(S,{name:"mesh-overview-view"},{default:_(({route:c})=>[e(g,{title:s(i)("meshes.routes.overview.title")},null,8,["title"]),d(),e(B,null,{default:_(()=>[e(D,{src:`/meshes/${c.params.mesh}`},{default:_(({data:a,error:m})=>[e(D,{src:`/mesh-insights/${c.params.mesh}`},{default:_(({data:f,error:v})=>[m?(t(),r(w,{key:0,error:m},null,8,["error"])):v?(t(),r(w,{key:1,error:v},null,8,["error"])):a===void 0||f===void 0?(t(),r(M,{key:2})):(t(),x("div",A,[e(s(o),{mesh:a,"mesh-insight":f},null,8,["mesh","mesh-insight"]),d(),R("div",I,[e(b,{"creation-time":a.creationTime,"modification-time":a.modificationTime},null,8,["creation-time","modification-time"])])]))]),_:2},1032,["src"])]),_:2},1032,["src"])]),_:2},1024)]),_:1}))}});const E=$(j,[["__scopeId","data-v-bc87e9d1"]]);export{E as default};
